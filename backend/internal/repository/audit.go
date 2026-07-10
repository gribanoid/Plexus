package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type AuditEventInput struct {
	ID           uuid.UUID
	OrgID        *uuid.UUID
	ActorID      uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	Metadata     map[string]any
}

func (r *Repo) InsertAuditEvent(ctx context.Context, in AuditEventInput) error {
	var metadataJSON []byte
	if len(in.Metadata) > 0 {
		var err error
		metadataJSON, err = json.Marshal(in.Metadata)
		if err != nil {
			return err
		}
	}

	sql, args, err := psql.Insert("audit_events").
		Columns("id", "org_id", "actor_id", "action", "resource_type", "resource_id", "metadata").
		Values(in.ID, in.OrgID, in.ActorID, in.Action, in.ResourceType, in.ResourceID, metadataJSON).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}
