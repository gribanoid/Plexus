package handler

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberws "github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/middleware"
	"github.com/plexus/backend/internal/storage"
	"github.com/plexus/backend/internal/websocket"
)

func (h *Handler) ListAttachments(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}

	items, err := h.Repo.ListAttachments(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"items": mapSlice(items, toAttachmentDTO)})
}

func (h *Handler) GetUploadURL(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}

	if _, err := h.Repo.GetIssueID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}

	var body struct {
		Filename string `json:"filename"`
		MimeType string `json:"mime_type"`
		Size     int64  `json:"size"`
	}
	if err := c.BodyParser(&body); err != nil || body.Filename == "" {
		return fiber.NewError(fiber.StatusBadRequest, "filename is required")
	}
	const maxUploadBytes int64 = 50 * 1024 * 1024
	if body.Size <= 0 || body.Size > maxUploadBytes {
		return fiber.NewError(fiber.StatusBadRequest, "size must be between 1 and 52428800 bytes")
	}
	safeName := filepath.Base(strings.ReplaceAll(body.Filename, "\\", "/"))
	if safeName == "" || safeName == "." || safeName == ".." {
		return fiber.NewError(fiber.StatusBadRequest, "invalid filename")
	}

	storageKey := fmt.Sprintf("uploads/%s/%s/%s", userID, uuid.New(), safeName)

	s3Client, err := h.s3Client(c.Context())
	if err != nil {
		return err
	}

	const expires = time.Hour
	uploadURL, err := s3Client.PresignPut(c.Context(), storageKey, expires)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"upload_url":  uploadURL,
		"storage_key": storageKey,
		"expires_in":  int(expires.Seconds()),
	})
}

func (h *Handler) CreateAttachment(c *fiber.Ctx) error {
	issueNumber, err := parseIssueNumber(c.Params("issueNumber"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue number")
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var body struct {
		Filename   string `json:"filename"`
		MimeType   string `json:"mime_type"`
		Size       int64  `json:"size"`
		StorageKey string `json:"storage_key"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	expectedPrefix := fmt.Sprintf("uploads/%s/", userID)
	if !strings.HasPrefix(body.StorageKey, expectedPrefix) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid storage_key")
	}

	issueID, err := h.Repo.GetIssueID(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"), issueNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "issue not found")
		}
		return err
	}

	id := uuid.New()
	if err := h.Repo.CreateAttachment(c.Context(), id, issueID, userID, body.Filename, body.MimeType, body.StorageKey, body.Size); err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) DeleteAttachment(c *fiber.Ctx) error {
	attachmentID, err := uuid.Parse(c.Params("attachmentID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid attachment id")
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	storageKey, n, err := h.Repo.DeleteAttachment(c.Context(), attachmentID, userID)
	if err != nil {
		return err
	}
	if n == 0 {
		return fiber.NewError(fiber.StatusForbidden, "not found or not owned by you")
	}

	if storageKey != "" {
		if s3Client, s3Err := h.s3Client(c.Context()); s3Err == nil {
			_ = s3Client.DeleteObject(c.Context(), storageKey)
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListNotifications(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	page, err := parsePageParams(c)
	if err != nil {
		return err
	}

	result, err := h.Repo.ListNotifications(c.Context(), userID, page)
	if err != nil {
		return err
	}
	return c.JSON(pageResponse(result, toNotificationDTO))
}

func (h *Handler) MarkNotificationRead(c *fiber.Ctx) error {
	notifID, err := uuid.Parse(c.Params("notificationID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid notification id")
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	if err := h.Repo.MarkNotificationRead(c.Context(), notifID, userID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) MarkAllNotificationsRead(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	if err := h.Repo.MarkAllNotificationsRead(c.Context(), userID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Search(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	orgSlug := c.Params("orgSlug")
	query := c.Query("q")
	projectKey := c.Query("project")

	if query == "" {
		return fiber.NewError(fiber.StatusBadRequest, "q is required")
	}

	var projectID string
	if projectKey != "" {
		id, err := h.Repo.GetProjectIDByOrgAndKey(c.Context(), userID, orgSlug, projectKey)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fiber.NewError(fiber.StatusNotFound, "project not found")
			}
			return err
		}
		ok, accessErr := h.Repo.UserHasProjectAccess(c.Context(), userID, id)
		if accessErr != nil {
			return accessErr
		}
		if !ok {
			return fiber.NewError(fiber.StatusForbidden, "access denied to project")
		}
		projectID = id.String()
	} else {
		// Org-wide search is limited to projects the user can access.
		ids, err := h.Repo.ListAccessibleProjectIDs(c.Context(), userID, orgSlug)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return c.JSON(fiber.Map{"items": []interface{}{}, "total": 0})
		}
		// Meilisearch client filters by a single project; search each and merge is heavy —
		// require project filter for now when multi-project.
		if len(ids) == 1 {
			projectID = ids[0].String()
		} else {
			return fiber.NewError(fiber.StatusBadRequest, "project query parameter is required")
		}
	}

	results, err := h.Deps.Search.Search(c.Context(), query, projectID, 25)
	if err != nil {
		return c.JSON(fiber.Map{"items": []interface{}{}, "total": 0})
	}
	return c.JSON(fiber.Map{"items": results.Hits, "total": results.Total})
}

func (h *Handler) WebSocketUpgrade(c *fiber.Ctx) error {
	if !fiberws.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	token := c.Query("token")
	if token == "" {
		header := c.Get("Authorization")
		if parts := strings.SplitN(header, " ", 2); len(parts) == 2 {
			token = parts[1]
		}
	}
	userID, err := middleware.ParseJWTToken(token, h.JWTSecret)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "project_id is required")
	}
	id, parseErr := uuid.Parse(projectIDStr)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project_id")
	}
	ok, accessErr := h.Repo.UserHasProjectAccess(c.Context(), userID, id)
	if accessErr != nil {
		return accessErr
	}
	if !ok {
		return fiber.NewError(fiber.StatusForbidden, "access denied to project")
	}
	projectID := id

	return fiberws.New(func(conn *fiberws.Conn) {
		client := &websocket.Client{
			UserID:    userID,
			ProjectID: &projectID,
			Send:      make(chan []byte, 256),
		}
		h.Hub.Register(client)

		go func() {
			for msg := range client.Send {
				if err := conn.WriteMessage(fiberws.TextMessage, msg); err != nil {
					break
				}
			}
		}()

		defer client.Close()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	})(c)
}

func (h *Handler) s3Client(ctx context.Context) (*storage.Client, error) {
	return storage.New(ctx, storage.Config{
		Endpoint:  h.S3Config.Endpoint,
		Bucket:    h.S3Config.Bucket,
		AccessKey: h.S3Config.AccessKey,
		SecretKey: h.S3Config.SecretKey,
		Region:    h.S3Config.Region,
	})
}
