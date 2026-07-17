package repository

import (
	"context"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type APIKeyDBO struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	CreatedBy uuid.UUID
	Scopes    []byte
	ExpiresAt *time.Time
	RevokedAt *time.Time
}

type APIKeyListDBO struct {
	ID         uuid.UUID
	Name       string
	Prefix     string
	Scopes     []byte
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	RevokedAt  *time.Time
	CreatedBy  uuid.UUID
	CreatedAt  time.Time
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

func (r *Repo) CreateAPIKey(ctx context.Context, id, orgID, createdBy uuid.UUID, name, keyHash, prefix string, scopes []string, expiresAt *time.Time) error {
	var scopesJSON []byte
	if scopes != nil {
		var err error
		scopesJSON, err = json.Marshal(scopes)
		if err != nil {
			return err
		}
	}
	sql, args, err := psql.Insert("api_keys").
		Columns("id", "org_id", "name", "key_hash", "prefix", "scopes", "expires_at", "created_by").
		Values(id, orgID, name, keyHash, prefix, scopesJSON, expiresAt, createdBy).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]APIKeyListDBO, error) {
	sql, args, err := psql.Select(
		"id", "name", "prefix", "scopes", "last_used_at", "expires_at", "revoked_at", "created_by", "created_at",
	).
		From("api_keys").
		Where(sq.Eq{"org_id": orgID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []APIKeyListDBO
	for rows.Next() {
		var k APIKeyListDBO
		if err := rows.Scan(&k.ID, &k.Name, &k.Prefix, &k.Scopes, &k.LastUsedAt, &k.ExpiresAt, &k.RevokedAt, &k.CreatedBy, &k.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, k)
	}
	return items, rows.Err()
}

func (r *Repo) RevokeAPIKey(ctx context.Context, id, orgID uuid.UUID) error {
	sql, args, err := psql.Update("api_keys").
		Set("revoked_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id, "org_id": orgID}).
		Where("revoked_at IS NULL").
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}
