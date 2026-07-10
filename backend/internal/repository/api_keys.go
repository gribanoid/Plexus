package repository

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type APIKeyDBO struct {
	ID         uuid.UUID
	OrgID      uuid.UUID
	CreatedBy  uuid.UUID
	Scopes     []byte
	ExpiresAt  *time.Time
	RevokedAt  *time.Time
}

func (r *Repo) GetAPIKeyByHash(ctx context.Context, keyHash string) (*APIKeyDBO, error) {
	sql, args, err := psql.Select("id", "org_id", "created_by", "scopes", "expires_at", "revoked_at").
		From("api_keys").
		Where(sq.Eq{"key_hash": keyHash}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var k APIKeyDBO
	var expiresAt, revokedAt *time.Time
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&k.ID, &k.OrgID, &k.CreatedBy, &k.Scopes, &expiresAt, &revokedAt)
	if err != nil {
		return nil, err
	}
	k.ExpiresAt = expiresAt
	k.RevokedAt = revokedAt
	return &k, nil
}

func (r *Repo) TouchAPIKeyLastUsed(ctx context.Context, keyID uuid.UUID) error {
	sql, args, err := psql.Update("api_keys").
		Set("last_used_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": keyID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}
