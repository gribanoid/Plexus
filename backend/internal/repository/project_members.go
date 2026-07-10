package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) ListProjectMembers(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) ([]ProjectMemberDBO, error) {
	sql, args, err := psql.Select(
		"u.id", "u.display_name", "u.email", "u.avatar_url", "pm.role", "pm.created_at",
	).
		From("project_members pm").
		Join("projects p ON p.id = pm.project_id").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Join("users u ON u.id = pm.user_id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey}).
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

	var members []ProjectMemberDBO
	for rows.Next() {
		var m ProjectMemberDBO
		if err := rows.Scan(&m.UserID, &m.DisplayName, &m.Email, &m.AvatarURL, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *Repo) UpsertProjectMember(ctx context.Context, projectID, userID uuid.UUID, role string) error {
	sql, args, err := psql.Insert("project_members").
		Columns("project_id", "user_id", "role").
		Values(projectID, userID, role).
		Suffix(`ON CONFLICT (project_id, user_id) DO UPDATE SET role = EXCLUDED.role, updated_at = NOW()`).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) DeleteProjectMember(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string, targetUserID uuid.UUID) error {
	sql, args, err := psql.Delete("project_members").
		Where(sq.Expr(`project_id = (
			SELECT p.id FROM projects p
			JOIN organizations o ON o.id = p.org_id
			JOIN org_members om ON om.org_id = o.id
			WHERE om.user_id = ? AND o.slug = ? AND p.key = ?
		)`, userID, orgSlug, projectKey)).
		Where(sq.Eq{"user_id": targetUserID}).
		ToSql()
	if err != nil {
		return err
	}
	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
