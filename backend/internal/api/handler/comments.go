package handler

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/middleware"
	"github.com/plexus/backend/internal/repository"
	"github.com/plexus/backend/internal/websocket"
)

func (h *Handler) ListComments(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}

	page, err := parsePageParams(c)
	if err != nil {
		return err
	}

	items, err := h.Repo.ListComments(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber, page)
	if err != nil {
		return err
	}
	return c.JSON(pageResponse(items, toCommentDTO))
}

func (h *Handler) CreateComment(c *fiber.Ctx) error {
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var body struct {
		Body string `json:"body"`
	}
	if err := c.BodyParser(&body); err != nil || body.Body == "" {
		return fiber.NewError(fiber.StatusBadRequest, "body is required")
	}

	issueID, projectID, err := h.Repo.ResolveIssue(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}

	commentID := uuid.New()
	if err := h.Repo.CreateComment(c.Context(), commentID, issueID, userID, body.Body); err != nil {
		return err
	}

	if meta, metaErr := h.Repo.GetIssueNotifyMeta(c.Context(), issueID); metaErr == nil {
		if meta.AssigneeID.Valid && meta.AssigneeID.V != userID {
			title := fmt.Sprintf("New comment on %s-%d", meta.ProjectKey, meta.Number)
			h.createNotification(c.Context(), repository.CreateNotificationInput{
				ID:      uuid.New(),
				UserID:  meta.AssigneeID.V,
				Type:    "commented",
				Title:   title,
				Body:    &body.Body,
				IssueID: &issueID,
			})
		}
	}

	h.Hub.Publish(&websocket.Event{
		Type:      websocket.EventCommentCreated,
		ProjectID: &projectID,
		Payload:   fiber.Map{"id": commentID, "issue_id": issueID},
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": commentID})
}

func (h *Handler) UpdateComment(c *fiber.Ctx) error {
	commentID, err := uuid.Parse(c.Params("commentID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid comment id")
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var body struct {
		Body string `json:"body"`
	}
	if err := c.BodyParser(&body); err != nil || body.Body == "" {
		return fiber.NewError(fiber.StatusBadRequest, "body is required")
	}

	n, err := h.Repo.UpdateComment(c.Context(), commentID, userID, body.Body)
	if err != nil {
		return err
	}
	if n == 0 {
		return fiber.NewError(fiber.StatusForbidden, "comment not found or not owned by you")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) DeleteComment(c *fiber.Ctx) error {
	commentID, err := uuid.Parse(c.Params("commentID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid comment id")
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	n, err := h.Repo.DeleteComment(c.Context(), commentID, userID)
	if err != nil {
		return err
	}
	if n == 0 {
		return fiber.NewError(fiber.StatusForbidden, "comment not found or not owned by you")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
