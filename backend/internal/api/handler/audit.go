package handler

import (
	"context"

	"github.com/google/uuid"
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
