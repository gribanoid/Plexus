package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) ListSprints(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) ([]SprintDBO, error) {
	sql, args, err := psql.Select("s.id", "s.name", "s.goal", "s.state", "s.start_date", "s.end_date", "s.created_at").
		From("sprints s").
		Join("projects p ON p.id = s.project_id").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey}).
		OrderBy("s.created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sprints []SprintDBO
	for rows.Next() {
		var s SprintDBO
		if err := rows.Scan(&s.ID, &s.Name, &s.Goal, &s.State, &s.StartDate, &s.EndDate, &s.CreatedAt); err != nil {
			return nil, err
		}
		sprints = append(sprints, s)
	}
	return sprints, rows.Err()
}

func (r *Repo) CreateSprint(ctx context.Context, id, projectID uuid.UUID, name string, goal *string, startDate, endDate interface{}) error {
	sql, args, err := psql.Insert("sprints").
		Columns("id", "project_id", "name", "goal", "start_date", "end_date").
		Values(id, projectID, name, goal, startDate, endDate).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) UpdateSprint(ctx context.Context, sprintID, projectID uuid.UUID, name, goal *string, startDate, endDate interface{}) error {
	q := psql.Update("sprints").Where(sq.Eq{"id": sprintID, "project_id": projectID})
	if name != nil {
		q = q.Set("name", *name)
	}
	if goal != nil {
		q = q.Set("goal", *goal)
	}
	if startDate != nil {
		q = q.Set("start_date", startDate)
	}
	if endDate != nil {
		q = q.Set("end_date", endDate)
	}
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) StartSprint(ctx context.Context, sprintID, projectID uuid.UUID) error {
	sql, args, err := psql.Update("sprints").
		Set("state", "active").
		Set("start_date", sq.Expr("COALESCE(start_date, NOW())")).
		Where(sq.Eq{"id": sprintID, "project_id": projectID, "state": "future"}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) CompleteSprint(ctx context.Context, sprintID, projectID uuid.UUID, moveToSprintID *uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql, args, err := psql.Update("sprints").
		Set("state", "closed").
		Set("end_date", sq.Expr("COALESCE(end_date, NOW())")).
		Where(sq.Eq{"id": sprintID, "project_id": projectID}).
		ToSql()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return err
	}

	if moveToSprintID != nil {
		sql, args, err = psql.Update("issues").
			Set("sprint_id", *moveToSprintID).
			Where(sq.Eq{"sprint_id": sprintID}).
			Where(sq.Expr(`status_id NOT IN (
				SELECT id FROM statuses WHERE project_id = ? AND category = 'done'
			)`, projectID)).
			ToSql()
	} else {
		sql, args, err = psql.Update("issues").
			Set("sprint_id", nil).
			Where(sq.Eq{"sprint_id": sprintID}).
			Where(sq.Expr(`status_id NOT IN (
				SELECT id FROM statuses WHERE project_id = ? AND category = 'done'
			)`, projectID)).
			ToSql()
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
