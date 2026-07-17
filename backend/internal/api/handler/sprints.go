package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/middleware"
	"github.com/plexus/backend/internal/websocket"
)

func (h *Handler) ListSprints(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	sprints, err := h.Repo.ListSprints(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"items": mapSlice(sprints, toSprintDTO)})
}

func (h *Handler) CreateSprint(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		Name      string     `json:"name"`
		Goal      *string    `json:"goal"`
		StartDate *time.Time `json:"start_date"`
		EndDate   *time.Time `json:"end_date"`
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
	if err := h.Repo.CreateSprint(c.Context(), id, projectID, body.Name, body.Goal, body.StartDate, body.EndDate); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateSprint(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	sprintID, err := uuid.Parse(c.Params("sprintID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid sprint id")
	}

	var body struct {
		Name      *string    `json:"name"`
		Goal      *string    `json:"goal"`
		StartDate *time.Time `json:"start_date"`
		EndDate   *time.Time `json:"end_date"`
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

	if err := h.Repo.UpdateSprint(c.Context(), sprintID, projectID, body.Name, body.Goal, body.StartDate, body.EndDate); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) StartSprint(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	sprintID, err := uuid.Parse(c.Params("sprintID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid sprint id")
	}

	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}

	if err := h.Repo.StartSprint(c.Context(), sprintID, projectID); err != nil {
		return err
	}

	h.Hub.Publish(&websocket.Event{
		Type:      websocket.EventSprintUpdated,
		ProjectID: &projectID,
		Payload:   fiber.Map{"sprint_id": sprintID, "state": "active"},
	})

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) CompleteSprint(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	sprintID, err := uuid.Parse(c.Params("sprintID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid sprint id")
	}

	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}

	var body struct {
		MoveUnfinishedToSprintID *uuid.UUID `json:"move_unfinished_to_sprint_id"`
	}
	_ = c.BodyParser(&body)

	if err := h.Repo.CompleteSprint(c.Context(), sprintID, projectID, body.MoveUnfinishedToSprintID); err != nil {
		return err
	}

	h.Hub.Publish(&websocket.Event{
		Type:      websocket.EventSprintUpdated,
		ProjectID: &projectID,
		Payload:   fiber.Map{"sprint_id": sprintID, "state": "closed"},
	})

	return c.SendStatus(fiber.StatusNoContent)
}
