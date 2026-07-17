package handler

import (
	"encoding/csv"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/middleware"
	"github.com/plexus/backend/internal/repository"
)

// Metrics exposes a minimal Prometheus text exposition for HA monitoring.
func Metrics(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/plain; version=0.0.4")
	return c.SendString("# HELP plexus_up 1 if the process is serving\n# TYPE plexus_up gauge\nplexus_up 1\n")
}

// SAMLMetadata returns IdP wiring instructions; native SAML is via identity broker in v1.
func (h *Handler) SAMLMetadata(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error":   "native SAML not enabled",
		"message": "Configure SAML via an OIDC-compatible identity broker (Keycloak, Authentik, Dex) and set OIDC_* env vars. See docs/ha-reference.md and docs/architecture.md.",
		"status":  "planned",
	})
}

// SAMLACS acknowledges ACS endpoint reservation for future native SAML.
func (h *Handler) SAMLACS(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error":  "native SAML ACS not enabled",
		"status": "planned",
	})
}

func (h *Handler) ProjectReportSummary(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	summary, err := h.Repo.ProjectIssueSummary(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	return c.JSON(summary)
}

func (h *Handler) ListEpicChildren(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}
	parentID, err := h.Repo.GetIssueID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}
	items, err := h.Repo.ListChildIssues(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), parentID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"items": mapSlice(items, toIssueListDTO)})
}

func (h *Handler) BulkUpdateIssues(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		IssueNumbers []int64     `json:"issue_numbers"`
		StatusID     *uuid.UUID  `json:"status_id"`
		AssigneeID   *uuid.UUID  `json:"assignee_id"`
		Priority     *string     `json:"priority"`
		SprintID     *uuid.UUID  `json:"sprint_id"`
		ClearSprint  bool        `json:"clear_sprint"`
	}
	if err := c.BodyParser(&body); err != nil || len(body.IssueNumbers) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "issue_numbers is required")
	}
	if len(body.IssueNumbers) > 100 {
		return fiber.NewError(fiber.StatusBadRequest, "max 100 issues per bulk update")
	}

	updated := 0
	for _, num := range body.IssueNumbers {
		issueID, _, err := h.Repo.ResolveIssue(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), num)
		if err != nil {
			continue
		}
		in := repository.UpdateIssueInput{
			StatusID:    body.StatusID,
			AssigneeID:  body.AssigneeID,
			Priority:    body.Priority,
			SprintID:    body.SprintID,
			ClearSprint: body.ClearSprint,
		}
		if err := h.Repo.UpdateIssue(c.Context(), issueID, in); err != nil {
			continue
		}
		updated++
	}
	return c.JSON(fiber.Map{"updated": updated})
}

func (h *Handler) ImportIssuesCSV(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "multipart file field 'file' is required")
	}
	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	statusID, err := h.Repo.FirstTodoStatusID(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "project has no statuses")
	}
	typeID, err := h.Repo.FirstIssueTypeID(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "project has no issue types")
	}

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true
	rows, err := r.ReadAll()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid CSV")
	}
	if len(rows) < 2 {
		return fiber.NewError(fiber.StatusBadRequest, "CSV must include header and at least one row")
	}

	header := make(map[string]int)
	for i, col := range rows[0] {
		header[strings.ToLower(strings.TrimSpace(col))] = i
	}
	titleIdx, ok := header["title"]
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "CSV must have a title column")
	}

	created := 0
	for _, row := range rows[1:] {
		if titleIdx >= len(row) {
			continue
		}
		title := strings.TrimSpace(row[titleIdx])
		if title == "" {
			continue
		}
		var description *string
		if idx, ok := header["description"]; ok && idx < len(row) && strings.TrimSpace(row[idx]) != "" {
			d := strings.TrimSpace(row[idx])
			description = &d
		}
		priority := "medium"
		if idx, ok := header["priority"]; ok && idx < len(row) && strings.TrimSpace(row[idx]) != "" {
			priority = strings.ToLower(strings.TrimSpace(row[idx]))
		}
		num, err := h.Repo.NextIssueNumber(c.Context(), projectID)
		if err != nil {
			return err
		}
		issueID := uuid.New()
		if err := h.Repo.CreateIssue(c.Context(), repository.CreateIssueInput{
			IssueID:     issueID,
			ProjectID:   projectID,
			Number:      num,
			TypeID:      typeID,
			StatusID:    statusID,
			Title:       title,
			Description: description,
			Priority:    priority,
			ReporterID:  userID,
			Position:    float64(time.Now().UnixNano()),
		}); err != nil {
			continue
		}
		created++
		_ = strconv.FormatInt(num, 10)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"created": created})
}
