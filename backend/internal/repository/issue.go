package repository

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) ListIssues(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string, filters IssueFilters, page PageParams) (PageResult[IssueListDBO], error) {
	q := psql.Select(
		"i.id", "i.number", "i.title", "i.priority", "i.story_points", "i.due_date", "i.position",
		"i.status_id", "i.type_id", "i.assignee_id", "i.reporter_id", "i.sprint_id",
		"i.created_at", "i.updated_at",
	).
		From("issues i").
		Join("projects p ON p.id = i.project_id").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey})

	if filters.StatusID != "" {
		q = q.Where(sq.Eq{"i.status_id": filters.StatusID})
	}
	if filters.AssigneeID != "" {
		q = q.Where(sq.Eq{"i.assignee_id": filters.AssigneeID})
	}
	if filters.SprintID != "" {
		q = q.Where(sq.Eq{"i.sprint_id": filters.SprintID})
	}
	if filters.Priority != "" {
		q = q.Where(sq.Eq{"i.priority": filters.Priority})
	}

	if page.Cursor != "" {
		cursorID, err := uuid.Parse(page.Cursor)
		if err != nil {
			return PageResult[IssueListDBO]{}, err
		}
		q = q.Where(sq.Or{
			sq.Expr("i.position > (SELECT position FROM issues WHERE id = ?)", cursorID),
			sq.And{
				sq.Expr("i.position = (SELECT position FROM issues WHERE id = ?)", cursorID),
				sq.Expr("i.created_at < (SELECT created_at FROM issues WHERE id = ?)", cursorID),
			},
			sq.And{
				sq.Expr("i.position = (SELECT position FROM issues WHERE id = ?)", cursorID),
				sq.Expr("i.created_at = (SELECT created_at FROM issues WHERE id = ?)", cursorID),
				sq.Expr("i.id > ?", cursorID),
			},
		})
	}

	limit := page.Limit
	if limit <= 0 {
		limit = 50
	}

	sql, args, err := q.OrderBy("i.position ASC", "i.created_at DESC", "i.id ASC").Limit(uint64(limit + 1)).ToSql()
	if err != nil {
		return PageResult[IssueListDBO]{}, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return PageResult[IssueListDBO]{}, err
	}
	defer rows.Close()

	var issues []IssueListDBO
	for rows.Next() {
		var i IssueListDBO
		if err := rows.Scan(
			&i.ID, &i.Number, &i.Title, &i.Priority, &i.StoryPoints, &i.DueDate, &i.Position,
			&i.StatusID, &i.TypeID, &i.AssigneeID, &i.ReporterID, &i.SprintID,
			&i.CreatedAt, &i.UpdatedAt,
		); err != nil {
			return PageResult[IssueListDBO]{}, err
		}
		issues = append(issues, i)
	}
	if err := rows.Err(); err != nil {
		return PageResult[IssueListDBO]{}, err
	}

	var nextCursor *string
	if len(issues) > limit {
		issues = issues[:limit]
		cursor := issues[len(issues)-1].ID.String()
		nextCursor = &cursor
	}
	return PageResult[IssueListDBO]{Items: issues, NextCursor: nextCursor}, nil
}

func (r *Repo) FirstTodoStatusID(ctx context.Context, projectID uuid.UUID) (uuid.UUID, error) {
	sql, args, err := psql.Select("id").
		From("statuses").
		Where(sq.Eq{"project_id": projectID, "category": "todo"}).
		OrderBy("position ASC").
		Limit(1).
		ToSql()
	if err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&id)
	return id, err
}

func (r *Repo) NextIssueNumber(ctx context.Context, projectID uuid.UUID) (int64, error) {
	sql, args, err := psql.Select("COALESCE(MAX(number), 0) + 1").
		From("issues").
		Where(sq.Eq{"project_id": projectID}).
		ToSql()
	if err != nil {
		return 0, err
	}
	var n int64
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&n)
	return n, err
}

func (r *Repo) MaxIssuePosition(ctx context.Context, projectID uuid.UUID) (float64, error) {
	sql, args, err := psql.Select("COALESCE(MAX(position), 0)").
		From("issues").
		Where(sq.Eq{"project_id": projectID}).
		ToSql()
	if err != nil {
		return 0, err
	}
	var pos float64
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&pos)
	return pos, err
}

func (r *Repo) CreateIssue(ctx context.Context, in CreateIssueInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql, args, err := psql.Insert("issues").
		Columns(
			"id", "project_id", "number", "type_id", "status_id", "title", "description",
			"priority", "assignee_id", "reporter_id", "parent_id", "sprint_id", "story_points", "due_date", "position",
		).
		Values(
			in.IssueID, in.ProjectID, in.Number, in.TypeID, in.StatusID, in.Title, in.Description,
			in.Priority, in.AssigneeID, in.ReporterID, in.ParentID, in.SprintID, in.StoryPoints, in.DueDate, in.Position,
		).
		ToSql()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return err
	}

	for _, labelID := range in.LabelIDs {
		sql, args, err = psql.Insert("issue_labels").
			Columns("issue_id", "label_id").
			Values(in.IssueID, labelID).
			ToSql()
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, sql, args...); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repo) GetIssue(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string, issueNumber int64) (*IssueDetailDBO, error) {
	sql, args, err := psql.Select(
		"i.id", "i.number", "i.title", "i.description", "i.priority", "i.story_points",
		"i.due_date", "i.position", "i.status_id", "i.type_id", "i.assignee_id",
		"assignee.display_name", "i.reporter_id", "reporter.display_name",
		"i.sprint_id", "i.parent_id", "i.created_at", "i.updated_at",
	).
		From("issues i").
		Join("projects p ON p.id = i.project_id").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		LeftJoin("users assignee ON assignee.id = i.assignee_id").
		LeftJoin("users reporter ON reporter.id = i.reporter_id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey, "i.number": issueNumber}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var issue IssueDetailDBO
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&issue.ID, &issue.Number, &issue.Title, &issue.Description, &issue.Priority,
		&issue.StoryPoints, &issue.DueDate, &issue.Position, &issue.StatusID, &issue.TypeID,
		&issue.AssigneeID, &issue.AssigneeName, &issue.ReporterID, &issue.ReporterName,
		&issue.SprintID, &issue.ParentID,
		&issue.CreatedAt, &issue.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

func (r *Repo) ResolveIssue(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string, issueNumber int64) (issueID, projectID uuid.UUID, err error) {
	sql, args, err := psql.Select("i.id", "p.id").
		From("issues i").
		Join("projects p ON p.id = i.project_id").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey, "i.number": issueNumber}).
		ToSql()
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&issueID, &projectID)
	return issueID, projectID, err
}

func (r *Repo) UpdateIssue(ctx context.Context, issueID uuid.UUID, in UpdateIssueInput) error {
	q := psql.Update("issues").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": issueID})
	if in.Title != nil {
		q = q.Set("title", *in.Title)
	}
	if in.Description != nil {
		q = q.Set("description", *in.Description)
	}
	if in.StatusID != nil {
		q = q.Set("status_id", *in.StatusID)
	}
	if in.TypeID != nil {
		q = q.Set("type_id", *in.TypeID)
	}
	if in.Priority != nil {
		q = q.Set("priority", sq.Expr("?::priority", *in.Priority))
	}
	if in.AssigneeID != nil {
		q = q.Set("assignee_id", *in.AssigneeID)
	}
	if in.SprintID != nil {
		q = q.Set("sprint_id", *in.SprintID)
	}
	if in.ParentID != nil {
		q = q.Set("parent_id", *in.ParentID)
	}
	if in.StoryPoints != nil {
		q = q.Set("story_points", *in.StoryPoints)
	}
	if in.DueDate != nil {
		q = q.Set("due_date", *in.DueDate)
	}
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) InsertIssueHistory(ctx context.Context, id, issueID, actorID uuid.UUID, field, oldValue, newValue string) error {
	sql, args, err := psql.Insert("issue_history").
		Columns("id", "issue_id", "actor_id", "field", "old_value", "new_value").
		Values(id, issueID, actorID, field, oldValue, newValue).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) DeleteIssue(ctx context.Context, issueID uuid.UUID) error {
	sql, args, err := psql.Delete("issues").Where(sq.Eq{"id": issueID}).ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) MoveIssue(ctx context.Context, issueID uuid.UUID, in MoveIssueInput) error {
	q := psql.Update("issues").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": issueID})
	if in.StatusID != nil {
		q = q.Set("status_id", *in.StatusID)
	}
	if in.Position != nil {
		q = q.Set("position", *in.Position)
	}
	if in.SprintID != nil {
		q = q.Set("sprint_id", *in.SprintID)
	}
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) ListIssueHistory(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string, issueNumber int64) ([]IssueHistoryDBO, error) {
	sql, args, err := psql.Select("ih.id", "ih.field", "ih.old_value", "ih.new_value", "ih.actor_id", "ih.created_at").
		From("issue_history ih").
		Join("issues i ON i.id = ih.issue_id").
		Join("projects p ON p.id = i.project_id").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey, "i.number": issueNumber}).
		OrderBy("ih.created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []IssueHistoryDBO
	for rows.Next() {
		var h IssueHistoryDBO
		if err := rows.Scan(&h.ID, &h.Field, &h.OldValue, &h.NewValue, &h.ActorID, &h.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, h)
	}
	return items, rows.Err()
}

func (r *Repo) ResolveIssueRef(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string, issueNumber int64) (issueID, projectID uuid.UUID, err error) {
	return r.ResolveIssue(ctx, userID, orgSlug, projectKey, issueNumber)
}

type IssueSearchDBO struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	Number       int64
	Title        string
	Description  sql.NullString
	Priority     string
	AssigneeName sql.NullString
	StatusName   string
	CreatedAt    time.Time
}

func (r *Repo) GetIssueForSearch(ctx context.Context, issueID uuid.UUID) (*IssueSearchDBO, error) {
	sql, args, err := psql.Select(
		"i.id", "i.project_id", "i.number", "i.title", "i.description", "i.priority",
		"u.display_name", "s.name", "i.created_at",
	).
		From("issues i").
		Join("statuses s ON s.id = i.status_id").
		LeftJoin("users u ON u.id = i.assignee_id").
		Where(sq.Eq{"i.id": issueID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var issue IssueSearchDBO
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&issue.ID, &issue.ProjectID, &issue.Number, &issue.Title, &issue.Description,
		&issue.Priority, &issue.AssigneeName, &issue.StatusName, &issue.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

type IssueNotifyMeta struct {
	AssigneeID sql.Null[uuid.UUID]
	Title      string
	Number     int64
	ProjectKey string
}

func (r *Repo) GetIssueNotifyMeta(ctx context.Context, issueID uuid.UUID) (*IssueNotifyMeta, error) {
	sql, args, err := psql.Select("i.assignee_id", "i.title", "i.number", "p.key").
		From("issues i").
		Join("projects p ON p.id = i.project_id").
		Where(sq.Eq{"i.id": issueID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var meta IssueNotifyMeta
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&meta.AssigneeID, &meta.Title, &meta.Number, &meta.ProjectKey)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}
