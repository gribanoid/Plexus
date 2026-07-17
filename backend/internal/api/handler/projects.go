package handler

import (
	"errors"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/middleware"
)

var projectKeyRegex = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`)

func (h *Handler) CreateProject(c *fiber.Ctx) error {
	orgSlug := c.Params("orgSlug")
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var body struct {
		Name        string  `json:"name"`
		Key         string  `json:"key"`
		Description *string `json:"description"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if body.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if body.Key == "" {
		body.Key = autoKey(body.Name)
	}
	if !projectKeyRegex.MatchString(body.Key) {
		return fiber.NewError(fiber.StatusBadRequest, "project key must be 2–10 uppercase letters/digits starting with a letter")
	}

	orgID, err := h.Repo.GetOrgIDForMember(c.Context(), userID, orgSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "organization not found or access denied")
		}
		return err
	}

	projectID := uuid.New()
	if err := h.Repo.CreateProject(c.Context(), projectID, orgID, userID, body.Key, body.Name, body.Description); err != nil {
		return fiber.NewError(fiber.StatusConflict, "project key already exists in this organization")
	}

	if err := h.Repo.UpsertProjectMember(c.Context(), projectID, userID, "admin"); err != nil {
		return err
	}

	h.logAudit(c.Context(), &orgID, userID, "create", "project", &projectID, map[string]any{
		"key":  body.Key,
		"name": body.Name,
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":          projectID,
		"org_id":      orgID,
		"key":         body.Key,
		"name":        body.Name,
		"description": body.Description,
	})
}

func (h *Handler) ListProjects(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	projects, err := h.Repo.ListProjects(c.Context(), userID, c.Params("orgSlug"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"items": mapSlice(projects, toProjectDTO)})
}

func (h *Handler) GetProject(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	p, err := h.Repo.GetProject(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	return c.JSON(toProjectDTO(*p))
}

func (h *Handler) UpdateProject(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var body struct {
		Name        *string    `json:"name"`
		Description *string    `json:"description"`
		IconURL     *string    `json:"icon_url"`
		LeadID      *uuid.UUID `json:"lead_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	err := h.Repo.UpdateProject(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"),
		body.Name, body.Description, body.IconURL, body.LeadID)
	if err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) DeleteProject(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	if admin, err := h.isOrgAdmin(c.Context(), c.Params("orgSlug"), userID); err != nil {
		return err
	} else if !admin {
		return fiber.NewError(fiber.StatusForbidden, "only org admins can delete projects")
	}

	if err := h.Repo.DeleteProject(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListStatuses(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	statuses, err := h.Repo.ListStatuses(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"items": mapSlice(statuses, toStatusDTO)})
}

func (h *Handler) CreateStatus(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		Name     string `json:"name"`
		Color    string `json:"color"`
		Category string `json:"category"`
		Position int    `json:"position"`
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
	if err := h.Repo.CreateStatus(c.Context(), id, projectID, body.Name,
		orDefault(body.Color, "#6B7280"), orDefault(body.Category, "todo"), body.Position); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateStatus(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	statusID, err := uuid.Parse(c.Params("statusID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid status id")
	}
	var body struct {
		Name     *string `json:"name"`
		Color    *string `json:"color"`
		Category *string `json:"category"`
		Position *int    `json:"position"`
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
	if err := h.Repo.UpdateStatus(c.Context(), statusID, projectID, body.Name, body.Color, body.Category, body.Position); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) DeleteStatus(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	statusID, err := uuid.Parse(c.Params("statusID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid status id")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	if err := h.Repo.DeleteStatus(c.Context(), statusID, projectID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListIssueTypes(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	items, err := h.Repo.ListIssueTypes(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"items": mapSlice(items, toIssueTypeDTO)})
}

func (h *Handler) CreateIssueType(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		Name    string  `json:"name"`
		Color   string  `json:"color"`
		IconURL *string `json:"icon_url"`
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
	if err := h.Repo.CreateIssueType(c.Context(), id, projectID, body.Name, orDefault(body.Color, "#6B7280"), body.IconURL); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) ListLabels(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	items, err := h.Repo.ListLabels(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"items": mapSlice(items, toLabelDTO)})
}

func (h *Handler) CreateLabel(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		Name  string `json:"name"`
		Color string `json:"color"`
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
	if err := h.Repo.CreateLabel(c.Context(), id, projectID, body.Name, orDefault(body.Color, "#6B7280")); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateLabel(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	labelID, err := uuid.Parse(c.Params("labelID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid label id")
	}
	var body struct {
		Name  *string `json:"name"`
		Color *string `json:"color"`
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
	if err := h.Repo.UpdateLabel(c.Context(), labelID, projectID, body.Name, body.Color); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) DeleteLabel(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	labelID, err := uuid.Parse(c.Params("labelID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid label id")
	}
	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}
	if err := h.Repo.DeleteLabel(c.Context(), labelID, projectID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func autoKey(name string) string {
	words := strings.Fields(name)
	key := ""
	for _, w := range words {
		if len(key) >= 4 {
			break
		}
		for _, ch := range w {
			if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' {
				key += strings.ToUpper(string(ch))
				break
			}
		}
	}
	if len(key) < 2 {
		key = "PRJ"
	}
	return key
}
