package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) GetOrgIDForMember(ctx context.Context, userID uuid.UUID, orgSlug string) (uuid.UUID, error) {
	sql, args, err := psql.Select("o.id").
		From("organizations o").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug}).
		ToSql()
	if err != nil {
		return uuid.Nil, err
	}
	var orgID uuid.UUID
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&orgID)
	return orgID, err
}

func (r *Repo) CreateProject(ctx context.Context, projectID, orgID, leadID uuid.UUID, key, name string, description *string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql, args, err := psql.Insert("projects").
		Columns("id", "org_id", "key", "name", "description", "lead_id").
		Values(projectID, orgID, key, name, description, leadID).
		ToSql()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "SELECT seed_project_defaults($1)", projectID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repo) ListProjects(ctx context.Context, userID uuid.UUID, orgSlug string) ([]ProjectDBO, error) {
	sql, args, err := psql.Select(
		"p.id", "p.key", "p.name", "p.description", "p.icon_url", "p.lead_id", "p.created_at",
	).
		From("projects p").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug}).
		OrderBy("p.name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []ProjectDBO
	for rows.Next() {
		var p ProjectDBO
		if err := rows.Scan(&p.ID, &p.Key, &p.Name, &p.Description, &p.IconURL, &p.LeadID, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *Repo) GetProject(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) (*ProjectDBO, error) {
	sql, args, err := psql.Select("p.id", "p.key", "p.name", "p.description", "p.icon_url", "p.lead_id").
		From("projects p").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey}).
		ToSql()
	if err != nil {
		return nil, err
	}
	var p ProjectDBO
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&p.ID, &p.Key, &p.Name, &p.Description, &p.IconURL, &p.LeadID,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) UpdateProject(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string, name, description, iconURL *string, leadID *uuid.UUID) error {
	q := psql.Update("projects").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"key": projectKey}).
		Where(sq.Expr(`org_id = (
			SELECT o.id FROM organizations o
			JOIN org_members om ON om.org_id = o.id AND om.user_id = ?
			WHERE o.slug = ?
		)`, userID, orgSlug))
	if name != nil {
		q = q.Set("name", *name)
	}
	if description != nil {
		q = q.Set("description", *description)
	}
	if iconURL != nil {
		q = q.Set("icon_url", *iconURL)
	}
	if leadID != nil {
		q = q.Set("lead_id", *leadID)
	}
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) DeleteProject(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) error {
	sql, args, err := psql.Delete("projects").
		Where(sq.Eq{"key": projectKey}).
		Where(sq.Expr(`org_id = (
			SELECT o.id FROM organizations o
			JOIN org_members om ON om.org_id = o.id AND om.user_id = ?
			WHERE o.slug = ?
		)`, userID, orgSlug)).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) ResolveProjectID(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) (uuid.UUID, error) {
	sql, args, err := psql.Select("p.id").
		From("projects p").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey}).
		ToSql()
	if err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&id)
	return id, err
}

func (r *Repo) UserHasProjectAccess(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	sql, args, err := psql.Select("1").
		From("projects p").
		Join("org_members om ON om.org_id = p.org_id").
		Where(sq.Eq{"p.id": projectID, "om.user_id": userID}).
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

func (r *Repo) ListStatuses(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) ([]StatusDBO, error) {
	sql, args, err := psql.Select("s.id", "s.name", "s.color", "s.category", "s.position").
		From("statuses s").
		Join("projects p ON p.id = s.project_id").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey}).
		OrderBy("s.position ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []StatusDBO
	for rows.Next() {
		var s StatusDBO
		if err := rows.Scan(&s.ID, &s.Name, &s.Color, &s.Category, &s.Position); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *Repo) CreateStatus(ctx context.Context, id, projectID uuid.UUID, name, color, category string, position int) error {
	sql, args, err := psql.Insert("statuses").
		Columns("id", "project_id", "name", "color", "category", "position").
		Values(id, projectID, name, color, category, position).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) UpdateStatus(ctx context.Context, statusID uuid.UUID, name, color, category *string, position *int) error {
	q := psql.Update("statuses").Where(sq.Eq{"id": statusID})
	if name != nil {
		q = q.Set("name", *name)
	}
	if color != nil {
		q = q.Set("color", *color)
	}
	if category != nil {
		q = q.Set("category", *category)
	}
	if position != nil {
		q = q.Set("position", *position)
	}
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) DeleteStatus(ctx context.Context, statusID uuid.UUID) error {
	sql, args, err := psql.Delete("statuses").Where(sq.Eq{"id": statusID}).ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) ListIssueTypes(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) ([]IssueTypeDBO, error) {
	sql, args, err := psql.Select("it.id", "it.name", "it.color", "it.icon_url").
		From("issue_types it").
		Join("projects p ON p.id = it.project_id").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey}).
		OrderBy("it.name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []IssueTypeDBO
	for rows.Next() {
		var it IssueTypeDBO
		if err := rows.Scan(&it.ID, &it.Name, &it.Color, &it.IconURL); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *Repo) CreateIssueType(ctx context.Context, id, projectID uuid.UUID, name, color string, iconURL *string) error {
	sql, args, err := psql.Insert("issue_types").
		Columns("id", "project_id", "name", "color", "icon_url").
		Values(id, projectID, name, color, iconURL).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) ListLabels(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) ([]LabelDBO, error) {
	sql, args, err := psql.Select("l.id", "l.name", "l.color").
		From("labels l").
		Join("projects p ON p.id = l.project_id").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey}).
		OrderBy("l.name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []LabelDBO
	for rows.Next() {
		var l LabelDBO
		if err := rows.Scan(&l.ID, &l.Name, &l.Color); err != nil {
			return nil, err
		}
		items = append(items, l)
	}
	return items, rows.Err()
}

func (r *Repo) CreateLabel(ctx context.Context, id, projectID uuid.UUID, name, color string) error {
	sql, args, err := psql.Insert("labels").
		Columns("id", "project_id", "name", "color").
		Values(id, projectID, name, color).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) GetProjectIDByOrgAndKey(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) (uuid.UUID, error) {
	return r.ResolveProjectID(ctx, userID, orgSlug, projectKey)
}
