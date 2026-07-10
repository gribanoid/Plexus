package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/jobs"
	"github.com/plexus/backend/internal/middleware"
	"github.com/plexus/backend/internal/repository"
	"github.com/plexus/backend/internal/websocket"
)

func parseIssueNumber(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func (h *Handler) ListIssues(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	page, err := parsePageParams(c)
	if err != nil {
		return err
	}

	issues, err := h.Repo.ListIssues(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), repository.IssueFilters{
		StatusID:   c.Query("status_id"),
		AssigneeID: c.Query("assignee_id"),
		SprintID:   c.Query("sprint_id"),
		Priority:   c.Query("priority"),
	}, page)
	if err != nil {
		return err
	}
	return c.JSON(pageResponse(issues, toIssueListDTO))
}

func (h *Handler) CreateIssue(c *fiber.Ctx) error {
	orgSlug := c.Params("orgSlug")
	projectKey := c.Params("projectKey")
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var body struct {
		Title       string      `json:"title"`
		TypeID      uuid.UUID   `json:"type_id"`
		StatusID    *uuid.UUID  `json:"status_id"`
		Priority    string      `json:"priority"`
		Description *string     `json:"description"`
		AssigneeID  *uuid.UUID  `json:"assignee_id"`
		ParentID    *uuid.UUID  `json:"parent_id"`
		SprintID    *uuid.UUID  `json:"sprint_id"`
		StoryPoints *float32   `json:"story_points"`
		DueDate     *time.Time `json:"due_date"`
		LabelIDs    []uuid.UUID `json:"label_ids"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if body.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}

	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, orgSlug, projectKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}

	statusID := body.StatusID
	if statusID == nil {
		first, err := h.Repo.FirstTodoStatusID(c.Context(), projectID)
		if err == nil {
			statusID = &first
		}
	}
	if statusID == nil {
		return fiber.NewError(fiber.StatusBadRequest, "status_id is required")
	}

	nextNumber, err := h.Repo.NextIssueNumber(c.Context(), projectID)
	if err != nil {
		return err
	}
	maxPosition, err := h.Repo.MaxIssuePosition(c.Context(), projectID)
	if err != nil {
		return err
	}

	issueID := uuid.New()
	if err := h.Repo.CreateIssue(c.Context(), repository.CreateIssueInput{
		IssueID:     issueID,
		ProjectID:   projectID,
		Number:      nextNumber,
		TypeID:      body.TypeID,
		StatusID:    *statusID,
		Title:       body.Title,
		Description: body.Description,
		Priority:    orDefault(body.Priority, "no_priority"),
		AssigneeID:  body.AssigneeID,
		ReporterID:  userID,
		ParentID:    body.ParentID,
		SprintID:    body.SprintID,
		StoryPoints: body.StoryPoints,
		DueDate:     body.DueDate,
		Position:    maxPosition + 65536,
		LabelIDs:    body.LabelIDs,
	}); err != nil {
		return err
	}

	payload, _ := json.Marshal(jobs.IndexIssuePayload{
		IssueID:   issueID.String(),
		ProjectID: projectID.String(),
		Title:     body.Title,
		Body:      ptrStr(body.Description),
	})
	_, _ = h.JobClient.Enqueue(asynq.NewTask(jobs.TaskIndexIssue, payload))

	h.Hub.Publish(&websocket.Event{
		Type:      websocket.EventIssueCreated,
		ProjectID: &projectID,
		Payload: fiber.Map{
			"id":     issueID,
			"number": nextNumber,
			"title":  body.Title,
		},
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":     issueID,
		"number": nextNumber,
	})
}

func (h *Handler) GetIssue(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}

	issue, err := h.Repo.GetIssue(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}

	dto := toIssueDetailDTO(*issue)
	if customValues, err := h.Repo.ListIssueCustomValues(c.Context(), issue.ID); err == nil {
		dto.CustomFields = customValuesToMap(customValues)
	}

	return c.JSON(dto)
}

func (h *Handler) UpdateIssue(c *fiber.Ctx) error {
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var body struct {
		Title        *string            `json:"title"`
		Description  *string            `json:"description"`
		StatusID     *uuid.UUID         `json:"status_id"`
		TypeID       *uuid.UUID         `json:"type_id"`
		Priority     *string            `json:"priority"`
		AssigneeID   *uuid.UUID         `json:"assignee_id"`
		SprintID     *uuid.UUID         `json:"sprint_id"`
		ParentID     *uuid.UUID         `json:"parent_id"`
		StoryPoints  *float32           `json:"story_points"`
		DueDate      *time.Time         `json:"due_date"`
		CustomFields map[string]*string `json:"custom_fields"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	issueID, projectID, err := h.Repo.ResolveIssue(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}

	issue, err := h.Repo.GetIssue(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}

	var prevAssigneeID *uuid.UUID
	if body.AssigneeID != nil && issue.AssigneeID.Valid {
		id := issue.AssigneeID.V
		prevAssigneeID = &id
	}

	updateInput := repository.UpdateIssueInput{
		Title:       body.Title,
		Description: body.Description,
		StatusID:    body.StatusID,
		TypeID:      body.TypeID,
		Priority:    body.Priority,
		AssigneeID:  body.AssigneeID,
		SprintID:    body.SprintID,
		ParentID:    body.ParentID,
		StoryPoints: body.StoryPoints,
		DueDate:     body.DueDate,
	}

	if err := h.Repo.UpdateIssue(c.Context(), issueID, updateInput); err != nil {
		return err
	}

	h.logIssueChanges(c.Context(), issueID, userID, *issue, updateInput)

	if len(body.CustomFields) > 0 {
		keys := make([]string, 0, len(body.CustomFields))
		for key := range body.CustomFields {
			keys = append(keys, key)
		}
		fieldIDs, err := h.Repo.ResolveCustomFieldIDs(c.Context(), projectID, keys)
		if err != nil {
			return err
		}
		upsert := make(map[uuid.UUID]*string, len(body.CustomFields))
		for key, value := range body.CustomFields {
			fieldID, ok := fieldIDs[key]
			if !ok {
				return fiber.NewError(fiber.StatusBadRequest, "unknown custom field: "+key)
			}
			upsert[fieldID] = value
		}
		if err := h.Repo.UpsertIssueCustomValues(c.Context(), issueID, upsert); err != nil {
			return err
		}
	}

	orgID, _ := h.Repo.GetOrgIDForMember(c.Context(), userID, c.Params("orgSlug"))
	h.logAudit(c.Context(), &orgID, userID, "update", "issue", &issueID, map[string]any{
		"project_key":  c.Params("projectKey"),
		"issue_number": issueNumber,
	})

	h.logIssueChanges(c.Context(), issueID, userID, *issue, updateInput)

	if body.AssigneeID != nil {
		newAssignee := *body.AssigneeID
		changed := prevAssigneeID == nil || *prevAssigneeID != newAssignee
		if changed && newAssignee != uuid.Nil && newAssignee != userID {
			projectKey := c.Params("projectKey")
			title := fmt.Sprintf("You were assigned to %s-%d", projectKey, issueNumber)
			h.createNotification(c.Context(), repository.CreateNotificationInput{
				ID:      uuid.New(),
				UserID:  newAssignee,
				Type:    "assigned",
				Title:   title,
				IssueID: &issueID,
			})
		}
	}

	title := issue.Title
	if body.Title != nil {
		title = *body.Title
	}
	description := nullStringVal(issue.Description)
	if body.Description != nil {
		description = *body.Description
	}
	h.enqueueIssueIndex(c.Context(), issueID, projectID, title, description)

	h.Hub.Publish(&websocket.Event{
		Type:      websocket.EventIssueUpdated,
		ProjectID: &projectID,
		Payload:   fiber.Map{"id": issueID, "issue_number": issueNumber},
	})

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) DeleteIssue(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}

	issueID, projectID, err := h.Repo.ResolveIssue(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}

	if err := h.Repo.DeleteIssue(c.Context(), issueID); err != nil {
		return err
	}

	payload, _ := json.Marshal(struct{ IssueID string `json:"issue_id"` }{IssueID: issueID.String()})
	_, _ = h.JobClient.Enqueue(asynq.NewTask(jobs.TaskDeleteIssueIndex, payload))

	h.Hub.Publish(&websocket.Event{
		Type:      websocket.EventIssueDeleted,
		ProjectID: &projectID,
		Payload:   fiber.Map{"id": issueID},
	})

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) MoveIssue(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}

	var body struct {
		StatusID *uuid.UUID `json:"status_id"`
		Position *float64   `json:"position"`
		SprintID *uuid.UUID `json:"sprint_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	issueID, projectID, err := h.Repo.ResolveIssue(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}

	issue, err := h.Repo.GetIssue(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}

	moveInput := repository.MoveIssueInput{
		StatusID: body.StatusID,
		Position: body.Position,
		SprintID: body.SprintID,
	}

	if err := h.Repo.MoveIssue(c.Context(), issueID, moveInput); err != nil {
		return err
	}

	h.logMoveIssueChanges(c.Context(), issueID, userID, *issue, moveInput)

	description := nullStringVal(issue.Description)
	h.enqueueIssueIndex(c.Context(), issueID, projectID, issue.Title, description)

	h.Hub.Publish(&websocket.Event{
		Type:      websocket.EventIssueUpdated,
		ProjectID: &projectID,
		Payload:   fiber.Map{"id": issueID, "moved": true},
	})

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) GetIssueHistory(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}

	items, err := h.Repo.ListIssueHistory(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"items": mapSlice(items, toIssueHistoryDTO)})
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func customValuesToMap(values []repository.IssueCustomValueDBO) map[string]*string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]*string, len(values))
	for _, v := range values {
		out[v.FieldKey] = v.Value
	}
	return out
}
