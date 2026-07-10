package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/middleware"
)

func (h *Handler) ListProjectMembers(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	members, err := h.Repo.ListProjectMembers(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"items": mapSlice(members, toProjectMemberDTO)})
}

func (h *Handler) AddProjectMember(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var body struct {
		UserID uuid.UUID `json:"user_id"`
		Role   string    `json:"role"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if body.UserID == uuid.Nil {
		return fiber.NewError(fiber.StatusBadRequest, "user_id is required")
	}
	if body.Role == "" {
		body.Role = "member"
	}

	projectID, err := h.Repo.ResolveProjectID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project not found")
		}
		return err
	}

	if err := h.Repo.UpsertProjectMember(c.Context(), projectID, body.UserID, body.Role); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) RemoveProjectMember(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	targetUserID, err := uuid.Parse(c.Params("userID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user id")
	}

	if err := h.Repo.DeleteProjectMember(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), targetUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "project member not found")
		}
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
