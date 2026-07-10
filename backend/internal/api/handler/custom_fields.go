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

var customFieldKeyRegex = regexp.MustCompile(`^[a-z][a-z0-9_]{1,49}$`)

func (h *Handler) ListCustomFields(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	items, err := h.Repo.ListCustomFields(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"items": mapSlice(items, toCustomFieldDTO)})
}

func (h *Handler) CreateCustomField(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var body struct {
		Name     string   `json:"name"`
		Key      string   `json:"key"`
		Type     string   `json:"type"`
		Required bool     `json:"required"`
		Options  []string `json:"options"`
		Position int      `json:"position"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if body.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if body.Key == "" {
		body.Key = customFieldKeyFromName(body.Name)
	}
	if !customFieldKeyRegex.MatchString(body.Key) {
		return fiber.NewError(fiber.StatusBadRequest, "key must be 2–50 lowercase letters, digits or underscores")
	}
	fieldType := orDefault(body.Type, "text")

	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}

	id := uuid.New()
	if err := h.Repo.CreateCustomField(c.Context(), id, projectID, body.Name, body.Key, fieldType, body.Required, body.Options, body.Position); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return fiber.NewError(fiber.StatusConflict, "custom field key already exists")
		}
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func customFieldKeyFromName(name string) string {
	s := strings.ToLower(name)
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if len(s) < 2 {
		s = "field"
	}
	if len(s) > 50 {
		s = s[:50]
	}
	return s
}
