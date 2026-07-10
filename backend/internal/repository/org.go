package repository

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) CreateOrg(ctx context.Context, orgID, userID uuid.UUID, slug, name string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql, args, err := psql.Insert("organizations").
		Columns("id", "slug", "name").
		Values(orgID, slug, name).
		ToSql()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return err
	}

	sql, args, err = psql.Insert("org_members").
		Columns("org_id", "user_id", "role").
		Values(orgID, userID, "owner").
		ToSql()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repo) ListMyOrgs(ctx context.Context, userID uuid.UUID) ([]OrgListDBO, error) {
	sql, args, err := psql.Select(
		"o.id", "o.slug", "o.name", "o.logo_url", "o.plan", "om.role", "o.created_at",
	).
		From("organizations o").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID}).
		OrderBy("o.name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []OrgListDBO
	for rows.Next() {
		var o OrgListDBO
		if err := rows.Scan(&o.ID, &o.Slug, &o.Name, &o.LogoURL, &o.Plan, &o.MyRole, &o.CreatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, o)
	}
	return orgs, rows.Err()
}

func (r *Repo) GetOrg(ctx context.Context, userID uuid.UUID, orgSlug string) (*OrgDetailDBO, error) {
	sql, args, err := psql.Select("o.id", "o.slug", "o.name", "o.logo_url", "o.plan", "om.role").
		From("organizations o").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"o.slug": orgSlug, "om.user_id": userID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var org OrgDetailDBO
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&org.ID, &org.Slug, &org.Name, &org.LogoURL, &org.Plan, &org.MyRole,
	)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *Repo) UpdateOrg(ctx context.Context, orgSlug string, name, logoURL *string) error {
	q := psql.Update("organizations").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"slug": orgSlug})
	if name != nil {
		q = q.Set("name", *name)
	}
	if logoURL != nil {
		q = q.Set("logo_url", *logoURL)
	}
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) ListOrgMembers(ctx context.Context, userID uuid.UUID, orgSlug string) ([]OrgMemberDBO, error) {
	sql, args, err := psql.Select(
		"u.id", "u.display_name", "u.email", "u.avatar_url", "om.role", "om.created_at",
	).
		From("org_members om").
		Join("users u ON u.id = om.user_id").
		Join("organizations o ON o.id = om.org_id").
		Join("org_members viewer ON viewer.org_id = o.id").
		Where(sq.Eq{"o.slug": orgSlug, "viewer.user_id": userID}).
		OrderBy("u.display_name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []OrgMemberDBO
	for rows.Next() {
		var m OrgMemberDBO
		if err := rows.Scan(&m.ID, &m.DisplayName, &m.Email, &m.AvatarURL, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *Repo) GetUserIDByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	sql, args, err := psql.Select("id").From("users").Where(sq.Eq{"email": email}).ToSql()
	if err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&id)
	return id, err
}

func (r *Repo) GetOrgIDBySlug(ctx context.Context, orgSlug string) (uuid.UUID, error) {
	sql, args, err := psql.Select("id").From("organizations").Where(sq.Eq{"slug": orgSlug}).ToSql()
	if err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&id)
	return id, err
}

func (r *Repo) UpsertOrgMember(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	sql, args, err := psql.Insert("org_members").
		Columns("org_id", "user_id", "role").
		Values(orgID, userID, role).
		Suffix("ON CONFLICT (org_id, user_id) DO UPDATE SET role = EXCLUDED.role").
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) RemoveOrgMember(ctx context.Context, orgSlug string, targetUserID uuid.UUID) error {
	sql, args, err := psql.Delete("org_members").
		Where(sq.Expr("org_id = (SELECT id FROM organizations WHERE slug = ?)", orgSlug)).
		Where(sq.Eq{"user_id": targetUserID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) IsOrgAdmin(ctx context.Context, orgSlug string, userID uuid.UUID) (bool, error) {
	sql, args, err := psql.Select("om.role").
		From("org_members om").
		Join("organizations o ON o.id = om.org_id").
		Where(sq.Eq{"o.slug": orgSlug, "om.user_id": userID}).
		ToSql()
	if err != nil {
		return false, err
	}
	var role string
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&role)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return role == "owner" || role == "admin", nil
}

func (r *Repo) OrgSlugExists(ctx context.Context, slug string) (bool, error) {
	sql, args, err := psql.Select("1").
		From("organizations").
		Where(sq.Eq{"slug": slug}).
		Limit(1).
		ToSql()
	if err != nil {
		return false, err
	}
	var one int
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&one)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *Repo) UniqueOrgSlug(ctx context.Context, base string) (string, error) {
	if base == "" {
		base = "workspace"
	}
	if len(base) < 3 {
		base = base + "-ws"
	}
	candidate := base
	for i := 0; i < 20; i++ {
		exists, err := r.OrgSlugExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i+1)
	}
	return base + "-" + uuid.New().String()[:8], nil
}
