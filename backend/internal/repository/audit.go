package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
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

type AuditEventDBO struct {
	ID           uuid.UUID
	OrgID        sql.Null[uuid.UUID]
	ActorID      sql.Null[uuid.UUID]
	Action       string
	ResourceType string
	ResourceID   sql.Null[uuid.UUID]
	Metadata     []byte
	CreatedAt    time.Time
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

	sqlStr, args, err := psql.Insert("audit_events").
		Columns("id", "org_id", "actor_id", "action", "resource_type", "resource_id", "metadata").
		Values(in.ID, in.OrgID, in.ActorID, in.Action, in.ResourceType, in.ResourceID, metadataJSON).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) ListAuditEvents(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]AuditEventDBO, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	sqlStr, args, err := psql.Select(
		"id", "org_id", "actor_id", "action", "resource_type", "resource_id", "metadata", "created_at",
	).
		From("audit_events").
		Where(sq.Eq{"org_id": orgID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AuditEventDBO
	for rows.Next() {
		var e AuditEventDBO
		if err := rows.Scan(&e.ID, &e.OrgID, &e.ActorID, &e.Action, &e.ResourceType, &e.ResourceID, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}
