package handler

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/middleware"
	"github.com/plexus/backend/internal/repository"
)

func (h *Handler) logAudit(ctx context.Context, orgID *uuid.UUID, actorID uuid.UUID, action, resourceType string, resourceID *uuid.UUID, metadata map[string]any) {
	_ = h.Repo.InsertAuditEvent(ctx, repository.AuditEventInput{
		ID:           uuid.New(),
		OrgID:        orgID,
		ActorID:      actorID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Metadata:     metadata,
	})
}

func (h *Handler) ListAuditEvents(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	if admin, err := h.isOrgAdmin(c.Context(), c.Params("orgSlug"), userID); err != nil {
		return err
	} else if !admin {
		return fiber.NewError(fiber.StatusForbidden, "only org admins can view audit events")
	}

	org, err := h.Repo.GetOrg(c.Context(), userID, c.Params("orgSlug"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "organization not found")
		}
		return err
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	items, err := h.Repo.ListAuditEvents(c.Context(), org.ID, limit, offset)
	if err != nil {
		return err
	}
	out := make([]fiber.Map, 0, len(items))
	for _, e := range items {
		var metadata any
		if len(e.Metadata) > 0 {
			_ = json.Unmarshal(e.Metadata, &metadata)
		}
		m := fiber.Map{
			"id":            e.ID,
			"action":        e.Action,
			"resource_type": e.ResourceType,
			"metadata":      metadata,
			"created_at":    e.CreatedAt,
		}
		if e.OrgID.Valid {
			m["org_id"] = e.OrgID.V
		}
		if e.ActorID.Valid {
			m["actor_id"] = e.ActorID.V
		}
		if e.ResourceID.Valid {
			m["resource_id"] = e.ResourceID.V
		}
		out = append(out, m)
	}
	return c.JSON(fiber.Map{"items": out})
}
