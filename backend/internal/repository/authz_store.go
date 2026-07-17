package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// GetOrgRole returns the organization ID and the caller's org membership role.
func (r *Repo) GetOrgRole(ctx context.Context, userID uuid.UUID, orgSlug string) (orgID uuid.UUID, role string, err error) {
	sql, args, err := psql.Select("o.id", "om.role").
		From("organizations o").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug}).
		ToSql()
	if err != nil {
		return uuid.Nil, "", err
	}
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&orgID, &role)
	return orgID, role, err
}

// GetProjectRole returns the caller's role in project_members.
func (r *Repo) GetProjectRole(ctx context.Context, userID, projectID uuid.UUID) (role string, err error) {
	sql, args, err := psql.Select("role").
		From("project_members").
		Where(sq.Eq{"user_id": userID, "project_id": projectID}).
		ToSql()
	if err != nil {
		return "", err
	}
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&role)
	return role, err
}

// ResolveProject looks up a project by org slug and key (no membership check).
func (r *Repo) ResolveProject(ctx context.Context, orgSlug, projectKey string) (projectID, orgID uuid.UUID, err error) {
	sql, args, err := psql.Select("p.id", "p.org_id").
		From("projects p").
		Join("organizations o ON o.id = p.org_id").
		Where(sq.Eq{"o.slug": orgSlug, "p.key": projectKey}).
		ToSql()
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&projectID, &orgID)
	return projectID, orgID, err
}
