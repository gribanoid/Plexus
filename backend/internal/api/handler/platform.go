package handler

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/crypto"
	"github.com/plexus/backend/internal/middleware"
	"github.com/plexus/backend/internal/safehttp"
)

// --- Workflow transitions ---

func (h *Handler) ListWorkflowTransitions(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	items, err := h.Repo.ListWorkflowTransitions(c.Context(), projectID)
	if err != nil {
		return err
	}
	out := make([]fiber.Map, 0, len(items))
	for _, t := range items {
		m := fiber.Map{
			"id":           t.ID,
			"project_id":   t.ProjectID,
			"to_status_id": t.ToStatusID,
			"name":         t.Name,
			"created_at":   t.CreatedAt,
		}
		if t.IssueTypeID.Valid {
			m["issue_type_id"] = t.IssueTypeID.V
		}
		if t.FromStatusID.Valid {
			m["from_status_id"] = t.FromStatusID.V
		}
		out = append(out, m)
	}
	return c.JSON(fiber.Map{"items": out})
}

func (h *Handler) CreateWorkflowTransition(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		IssueTypeID  *uuid.UUID `json:"issue_type_id"`
		FromStatusID *uuid.UUID `json:"from_status_id"`
		ToStatusID   uuid.UUID  `json:"to_status_id"`
		Name         string     `json:"name"`
	}
	if err := c.BodyParser(&body); err != nil || body.ToStatusID == uuid.Nil {
		return fiber.NewError(fiber.StatusBadRequest, "to_status_id is required")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	id := uuid.New()
	if err := h.Repo.CreateWorkflowTransition(c.Context(), id, projectID, body.IssueTypeID, body.FromStatusID, body.ToStatusID, body.Name); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) DeleteWorkflowTransition(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	transitionID, err := uuid.Parse(c.Params("transitionID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid transition id")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	if err := h.Repo.DeleteWorkflowTransition(c.Context(), transitionID, projectID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// --- Issue links ---

func (h *Handler) ListIssueLinks(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}
	issueID, err := h.Repo.GetIssueID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}
	items, err := h.Repo.ListIssueLinks(c.Context(), issueID)
	if err != nil {
		return err
	}
	out := make([]fiber.Map, 0, len(items))
	for _, l := range items {
		m := fiber.Map{
			"id":         l.ID,
			"source_id":  l.SourceID,
			"target_id":  l.TargetID,
			"link_type":  l.LinkType,
			"created_at": l.CreatedAt,
		}
		if l.CreatedBy.Valid {
			m["created_by"] = l.CreatedBy.V
		}
		out = append(out, m)
	}
	return c.JSON(fiber.Map{"items": out})
}

func (h *Handler) CreateIssueLink(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}
	var body struct {
		TargetID uuid.UUID `json:"target_id"`
		LinkType string    `json:"link_type"`
	}
	if err := c.BodyParser(&body); err != nil || body.TargetID == uuid.Nil {
		return fiber.NewError(fiber.StatusBadRequest, "target_id is required")
	}
	if body.LinkType == "" {
		body.LinkType = "relates_to"
	}
	sourceID, err := h.Repo.GetIssueID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	targetProjectID, err := h.Repo.GetIssueProjectID(c.Context(), body.TargetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "target issue not found")
		}
		return err
	}
	if targetProjectID != projectID {
		return fiber.NewError(fiber.StatusBadRequest, "target issue must belong to the same project")
	}
	id := uuid.New()
	if err := h.Repo.CreateIssueLink(c.Context(), id, sourceID, body.TargetID, body.LinkType, &userID); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) DeleteIssueLink(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	linkID, err := uuid.Parse(c.Params("linkID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid link id")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	if err := h.Repo.DeleteIssueLink(c.Context(), linkID, projectID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// --- Saved filters ---

func (h *Handler) ListSavedFilters(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	items, err := h.Repo.ListSavedFilters(c.Context(), projectID, userID)
	if err != nil {
		return err
	}
	out := make([]fiber.Map, 0, len(items))
	for _, f := range items {
		var query any
		_ = json.Unmarshal(f.Query, &query)
		out = append(out, fiber.Map{
			"id":         f.ID,
			"project_id": f.ProjectID,
			"owner_id":   f.OwnerID,
			"name":       f.Name,
			"query":      query,
			"is_shared":  f.IsShared,
			"created_at": f.CreatedAt,
			"updated_at": f.UpdatedAt,
		})
	}
	return c.JSON(fiber.Map{"items": out})
}

func (h *Handler) CreateSavedFilter(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		Name     string          `json:"name"`
		Query    json.RawMessage `json:"query"`
		IsShared bool            `json:"is_shared"`
	}
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	id := uuid.New()
	if err := h.Repo.CreateSavedFilter(c.Context(), id, projectID, userID, body.Name, body.Query, body.IsShared); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateSavedFilter(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	filterID, err := uuid.Parse(c.Params("filterID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid filter id")
	}
	var body struct {
		Name     *string         `json:"name"`
		Query    json.RawMessage `json:"query"`
		IsShared *bool           `json:"is_shared"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := h.Repo.UpdateSavedFilter(c.Context(), filterID, userID, body.Name, body.Query, body.IsShared); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) DeleteSavedFilter(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	filterID, err := uuid.Parse(c.Params("filterID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid filter id")
	}
	if err := h.Repo.DeleteSavedFilter(c.Context(), filterID, userID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// --- Versions ---

func (h *Handler) ListVersions(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	items, err := h.Repo.ListVersions(c.Context(), projectID)
	if err != nil {
		return err
	}
	out := make([]fiber.Map, 0, len(items))
	for _, v := range items {
		m := fiber.Map{
			"id":         v.ID,
			"project_id": v.ProjectID,
			"name":       v.Name,
			"status":     v.Status,
			"created_at": v.CreatedAt,
			"updated_at": v.UpdatedAt,
		}
		if v.Description.Valid {
			m["description"] = v.Description.String
		}
		if v.StartDate.Valid {
			m["start_date"] = v.StartDate.Time
		}
		if v.ReleaseDate.Valid {
			m["release_date"] = v.ReleaseDate.Time
		}
		out = append(out, m)
	}
	return c.JSON(fiber.Map{"items": out})
}

func (h *Handler) CreateVersion(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		Name        string     `json:"name"`
		Description *string    `json:"description"`
		Status      string     `json:"status"`
		StartDate   *time.Time `json:"start_date"`
		ReleaseDate *time.Time `json:"release_date"`
	}
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	id := uuid.New()
	if err := h.Repo.CreateVersion(c.Context(), id, projectID, body.Name, body.Description, body.Status, body.StartDate, body.ReleaseDate); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateVersion(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	versionID, err := uuid.Parse(c.Params("versionID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid version id")
	}
	var body struct {
		Name        *string    `json:"name"`
		Description *string    `json:"description"`
		Status      *string    `json:"status"`
		StartDate   *time.Time `json:"start_date"`
		ReleaseDate *time.Time `json:"release_date"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	if err := h.Repo.UpdateVersion(c.Context(), versionID, projectID, body.Name, body.Description, body.Status, body.StartDate, body.ReleaseDate); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) DeleteVersion(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	versionID, err := uuid.Parse(c.Params("versionID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid version id")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	if err := h.Repo.DeleteVersion(c.Context(), versionID, projectID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) SetIssueVersions(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}
	var body struct {
		VersionIDs []uuid.UUID `json:"version_ids"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	issueID, err := h.Repo.GetIssueID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}
	if err := h.Repo.SetIssueVersions(c.Context(), issueID, body.VersionIDs); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// --- Components ---

func (h *Handler) ListComponents(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	items, err := h.Repo.ListComponents(c.Context(), projectID)
	if err != nil {
		return err
	}
	out := make([]fiber.Map, 0, len(items))
	for _, comp := range items {
		m := fiber.Map{
			"id":         comp.ID,
			"project_id": comp.ProjectID,
			"name":       comp.Name,
			"created_at": comp.CreatedAt,
		}
		if comp.Description.Valid {
			m["description"] = comp.Description.String
		}
		if comp.LeadID.Valid {
			m["lead_id"] = comp.LeadID.V
		}
		out = append(out, m)
	}
	return c.JSON(fiber.Map{"items": out})
}

func (h *Handler) CreateComponent(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		Name        string     `json:"name"`
		Description *string    `json:"description"`
		LeadID      *uuid.UUID `json:"lead_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	id := uuid.New()
	if err := h.Repo.CreateComponent(c.Context(), id, projectID, body.Name, body.Description, body.LeadID); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) DeleteComponent(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	componentID, err := uuid.Parse(c.Params("componentID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid component id")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	if err := h.Repo.DeleteComponent(c.Context(), componentID, projectID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) SetIssueComponents(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}
	var body struct {
		ComponentIDs []uuid.UUID `json:"component_ids"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	issueID, err := h.Repo.GetIssueID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}
	if err := h.Repo.SetIssueComponents(c.Context(), issueID, body.ComponentIDs); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// --- Watchers ---

func (h *Handler) ListWatchers(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}
	issueID, err := h.Repo.GetIssueID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}
	items, err := h.Repo.ListWatchers(c.Context(), issueID)
	if err != nil {
		return err
	}
	out := make([]fiber.Map, 0, len(items))
	for _, w := range items {
		m := fiber.Map{
			"user_id":      w.UserID,
			"display_name": w.DisplayName,
			"email":        w.Email,
			"created_at":   w.CreatedAt,
		}
		if w.AvatarURL.Valid {
			m["avatar_url"] = w.AvatarURL.String
		}
		out = append(out, m)
	}
	return c.JSON(fiber.Map{"items": out})
}

func (h *Handler) AddWatcher(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}
	var body struct {
		UserID *uuid.UUID `json:"user_id"`
	}
	_ = c.BodyParser(&body)
	watcherID := userID
	if body.UserID != nil {
		watcherID = *body.UserID
	}
	issueID, err := h.Repo.GetIssueID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}
	if err := h.Repo.AddWatcher(c.Context(), issueID, watcherID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) RemoveWatcher(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}
	watcherID, err := uuid.Parse(c.Params("userID"))
	if err != nil {
		watcherID = userID
	}
	issueID, err := h.Repo.GetIssueID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}
	if err := h.Repo.RemoveWatcher(c.Context(), issueID, watcherID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// --- Webhooks ---

func (h *Handler) ListWebhooks(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	org, err := h.Repo.GetOrg(c.Context(), userID, c.Params("orgSlug"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "organization not found")
		}
		return err
	}
	items, err := h.Repo.ListWebhooks(c.Context(), org.ID)
	if err != nil {
		return err
	}
	out := make([]fiber.Map, 0, len(items))
	for _, w := range items {
		var events any
		_ = json.Unmarshal(w.Events, &events)
		m := fiber.Map{
			"id":         w.ID,
			"org_id":     w.OrgID,
			"name":       w.Name,
			"url":        w.URL,
			"events":     events,
			"active":     w.Active,
			"created_at": w.CreatedAt,
			"updated_at": w.UpdatedAt,
		}
		if w.CreatedBy.Valid {
			m["created_by"] = w.CreatedBy.V
		}
		out = append(out, m)
	}
	return c.JSON(fiber.Map{"items": out})
}

func (h *Handler) CreateWebhook(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		Name   string   `json:"name"`
		URL    string   `json:"url"`
		Secret string   `json:"secret"`
		Events []string `json:"events"`
	}
	if err := c.BodyParser(&body); err != nil || body.Name == "" || body.URL == "" || body.Secret == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name, url and secret are required")
	}
	if err := safehttp.ValidateWebhookURL(body.URL); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	org, err := h.Repo.GetOrg(c.Context(), userID, c.Params("orgSlug"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "organization not found")
		}
		return err
	}
	id := uuid.New()
	encSecret, err := crypto.EncryptString(h.EncryptionKey, body.Secret)
	if err != nil {
		return err
	}
	if err := h.Repo.CreateWebhook(c.Context(), id, org.ID, body.Name, body.URL, encSecret, body.Events, &userID); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateWebhook(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	webhookID, err := uuid.Parse(c.Params("webhookID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid webhook id")
	}
	var body struct {
		Name   *string  `json:"name"`
		URL    *string  `json:"url"`
		Secret *string  `json:"secret"`
		Events []string `json:"events"`
		Active *bool    `json:"active"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if body.URL != nil {
		if err := safehttp.ValidateWebhookURL(*body.URL); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}
	var encSecret *string
	if body.Secret != nil {
		enc, err := crypto.EncryptString(h.EncryptionKey, *body.Secret)
		if err != nil {
			return err
		}
		encSecret = &enc
	}
	org, err := h.Repo.GetOrg(c.Context(), userID, c.Params("orgSlug"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "organization not found")
		}
		return err
	}
	if err := h.Repo.UpdateWebhook(c.Context(), webhookID, org.ID, body.Name, body.URL, encSecret, body.Events, body.Active); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) DeleteWebhook(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	webhookID, err := uuid.Parse(c.Params("webhookID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid webhook id")
	}
	org, err := h.Repo.GetOrg(c.Context(), userID, c.Params("orgSlug"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "organization not found")
		}
		return err
	}
	if err := h.Repo.DeleteWebhook(c.Context(), webhookID, org.ID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// --- Automation rules ---

func (h *Handler) ListAutomationRules(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	items, err := h.Repo.ListAutomationRules(c.Context(), projectID)
	if err != nil {
		return err
	}
	out := make([]fiber.Map, 0, len(items))
	for _, a := range items {
		var conditions, actions any
		_ = json.Unmarshal(a.Conditions, &conditions)
		_ = json.Unmarshal(a.Actions, &actions)
		m := fiber.Map{
			"id":         a.ID,
			"project_id": a.ProjectID,
			"name":       a.Name,
			"enabled":    a.Enabled,
			"trigger":    a.Trigger,
			"conditions": conditions,
			"actions":    actions,
			"created_at": a.CreatedAt,
			"updated_at": a.UpdatedAt,
		}
		if a.CreatedBy.Valid {
			m["created_by"] = a.CreatedBy.V
		}
		out = append(out, m)
	}
	return c.JSON(fiber.Map{"items": out})
}

func (h *Handler) CreateAutomationRule(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		Name       string          `json:"name"`
		Trigger    string          `json:"trigger"`
		Conditions json.RawMessage `json:"conditions"`
		Actions    json.RawMessage `json:"actions"`
	}
	if err := c.BodyParser(&body); err != nil || body.Name == "" || body.Trigger == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name and trigger are required")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	id := uuid.New()
	if err := h.Repo.CreateAutomationRule(c.Context(), id, projectID, body.Name, body.Trigger, body.Conditions, body.Actions, &userID); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateAutomationRule(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	ruleID, err := uuid.Parse(c.Params("ruleID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid rule id")
	}
	var body struct {
		Name       *string         `json:"name"`
		Trigger    *string         `json:"trigger"`
		Enabled    *bool           `json:"enabled"`
		Conditions json.RawMessage `json:"conditions"`
		Actions    json.RawMessage `json:"actions"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	if err := h.Repo.UpdateAutomationRule(c.Context(), ruleID, projectID, body.Name, body.Trigger, body.Enabled, body.Conditions, body.Actions); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) DeleteAutomationRule(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	ruleID, err := uuid.Parse(c.Params("ruleID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid rule id")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	if err := h.Repo.DeleteAutomationRule(c.Context(), ruleID, projectID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// --- Permission schemes ---

func (h *Handler) ListPermissionSchemes(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	org, err := h.Repo.GetOrg(c.Context(), userID, c.Params("orgSlug"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "organization not found")
		}
		return err
	}
	items, err := h.Repo.ListPermissionSchemes(c.Context(), org.ID)
	if err != nil {
		return err
	}
	out := make([]fiber.Map, 0, len(items))
	for _, p := range items {
		var grants any
		_ = json.Unmarshal(p.Grants, &grants)
		m := fiber.Map{
			"id":         p.ID,
			"org_id":     p.OrgID,
			"name":       p.Name,
			"grants":     grants,
			"created_at": p.CreatedAt,
			"updated_at": p.UpdatedAt,
		}
		if p.Description.Valid {
			m["description"] = p.Description.String
		}
		out = append(out, m)
	}
	return c.JSON(fiber.Map{"items": out})
}

func (h *Handler) CreatePermissionScheme(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		Name        string          `json:"name"`
		Description *string         `json:"description"`
		Grants      json.RawMessage `json:"grants"`
	}
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	org, err := h.Repo.GetOrg(c.Context(), userID, c.Params("orgSlug"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "organization not found")
		}
		return err
	}
	id := uuid.New()
	if err := h.Repo.CreatePermissionScheme(c.Context(), id, org.ID, body.Name, body.Description, body.Grants); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdatePermissionScheme(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	schemeID, err := uuid.Parse(c.Params("schemeID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid scheme id")
	}
	var body struct {
		Name        *string         `json:"name"`
		Description *string         `json:"description"`
		Grants      json.RawMessage `json:"grants"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	org, err := h.Repo.GetOrg(c.Context(), userID, c.Params("orgSlug"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "organization not found")
		}
		return err
	}
	if err := h.Repo.UpdatePermissionScheme(c.Context(), schemeID, org.ID, body.Name, body.Description, body.Grants); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) AssignPermissionScheme(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		SchemeID *uuid.UUID `json:"scheme_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	if err := h.Repo.AssignPermissionScheme(c.Context(), projectID, body.SchemeID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// --- Soft-delete issue ---

func (h *Handler) SoftDeleteIssue(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}
	issueID, err := h.Repo.GetIssueID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}
	now := time.Now().UTC()
	if err := h.Repo.SetIssueDeletedAt(c.Context(), issueID, &now); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) RestoreIssue(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}
	// Resolve without deleted_at filter via direct lookup by number+project
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	issueID, err := h.Repo.GetIssueIDIncludingDeleted(c.Context(), projectID, issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}
	if err := h.Repo.SetIssueDeletedAt(c.Context(), issueID, nil); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
