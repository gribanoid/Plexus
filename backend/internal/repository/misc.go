package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) ListComments(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string, issueNumber int64, page PageParams) (PageResult[CommentDBO], error) {
	q := psql.Select("cm.id", "cm.body", "cm.author_id", "cm.created_at", "cm.updated_at").
		From("comments cm").
		Join("issues i ON i.id = cm.issue_id").
		Join("projects p ON p.id = i.project_id").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey, "i.number": issueNumber})

	if page.Cursor != "" {
		cursorID, err := uuid.Parse(page.Cursor)
		if err != nil {
			return PageResult[CommentDBO]{}, err
		}
		q = q.Where(sq.Or{
			sq.Expr("cm.created_at > (SELECT created_at FROM comments WHERE id = ?)", cursorID),
			sq.And{
				sq.Expr("cm.created_at = (SELECT created_at FROM comments WHERE id = ?)", cursorID),
				sq.Expr("cm.id > ?", cursorID),
			},
		})
	}

	limit := page.Limit
	if limit <= 0 {
		limit = 50
	}

	sql, args, err := q.OrderBy("cm.created_at ASC", "cm.id ASC").Limit(uint64(limit + 1)).ToSql()
	if err != nil {
		return PageResult[CommentDBO]{}, err
	}
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return PageResult[CommentDBO]{}, err
	}
	defer rows.Close()

	var items []CommentDBO
	for rows.Next() {
		var cm CommentDBO
		if err := rows.Scan(&cm.ID, &cm.Body, &cm.AuthorID, &cm.CreatedAt, &cm.UpdatedAt); err != nil {
			return PageResult[CommentDBO]{}, err
		}
		items = append(items, cm)
	}
	if err := rows.Err(); err != nil {
		return PageResult[CommentDBO]{}, err
	}

	var nextCursor *string
	if len(items) > limit {
		items = items[:limit]
		cursor := items[len(items)-1].ID.String()
		nextCursor = &cursor
	}
	return PageResult[CommentDBO]{Items: items, NextCursor: nextCursor}, nil
}

func (r *Repo) CreateComment(ctx context.Context, id, issueID, authorID uuid.UUID, body string) error {
	sql, args, err := psql.Insert("comments").
		Columns("id", "issue_id", "author_id", "body").
		Values(id, issueID, authorID, body).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) UpdateComment(ctx context.Context, commentID, authorID uuid.UUID, body string) (int64, error) {
	sql, args, err := psql.Update("comments").
		Set("body", body).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": commentID, "author_id": authorID}).
		ToSql()
	if err != nil {
		return 0, err
	}
	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *Repo) DeleteComment(ctx context.Context, commentID, authorID uuid.UUID) (int64, error) {
	sql, args, err := psql.Delete("comments").
		Where(sq.Eq{"id": commentID, "author_id": authorID}).
		ToSql()
	if err != nil {
		return 0, err
	}
	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *Repo) ListAttachments(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string, issueNumber int64) ([]AttachmentDBO, error) {
	sql, args, err := psql.Select("a.id", "a.filename", "a.mime_type", "a.size", "a.uploader_id", "a.created_at").
		From("attachments a").
		Join("issues i ON i.id = a.issue_id").
		Join("projects p ON p.id = i.project_id").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey, "i.number": issueNumber}).
		OrderBy("a.created_at ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AttachmentDBO
	for rows.Next() {
		var a AttachmentDBO
		if err := rows.Scan(&a.ID, &a.Filename, &a.MimeType, &a.Size, &a.UploaderID, &a.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

func (r *Repo) GetIssueID(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string, issueNumber int64) (uuid.UUID, error) {
	sql, args, err := psql.Select("i.id").
		From("issues i").
		Join("projects p ON p.id = i.project_id").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey, "i.number": issueNumber}).
		Where("i.deleted_at IS NULL").
		ToSql()
	if err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&id)
	return id, err
}

func (r *Repo) GetIssueProjectID(ctx context.Context, issueID uuid.UUID) (uuid.UUID, error) {
	sql, args, err := psql.Select("project_id").
		From("issues").
		Where(sq.Eq{"id": issueID}).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return uuid.Nil, err
	}
	var projectID uuid.UUID
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&projectID)
	return projectID, err
}

func (r *Repo) CreateAttachment(ctx context.Context, id, issueID, uploaderID uuid.UUID, filename, mimeType, storageKey string, size int64) error {
	sql, args, err := psql.Insert("attachments").
		Columns("id", "issue_id", "uploader_id", "filename", "mime_type", "size", "storage_key").
		Values(id, issueID, uploaderID, filename, mimeType, size, storageKey).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) DeleteAttachment(ctx context.Context, attachmentID, uploaderID uuid.UUID) (storageKey string, n int64, err error) {
	sql, args, err := psql.Delete("attachments").
		Where(sq.Eq{"id": attachmentID, "uploader_id": uploaderID}).
		Suffix("RETURNING storage_key").
		ToSql()
	if err != nil {
		return "", 0, err
	}
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&storageKey)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", 0, nil
		}
		return "", 0, err
	}
	return storageKey, 1, nil
}

func (r *Repo) ListNotifications(ctx context.Context, userID uuid.UUID, page PageParams) (PageResult[NotificationDBO], error) {
	q := psql.Select("id", "type", "title", "body", "read", "issue_id", "created_at").
		From("notifications").
		Where(sq.Eq{"user_id": userID})

	if page.Cursor != "" {
		cursorID, err := uuid.Parse(page.Cursor)
		if err != nil {
			return PageResult[NotificationDBO]{}, err
		}
		q = q.Where(sq.Or{
			sq.Expr("created_at < (SELECT created_at FROM notifications WHERE id = ?)", cursorID),
			sq.And{
				sq.Expr("created_at = (SELECT created_at FROM notifications WHERE id = ?)", cursorID),
				sq.Expr("id < ?", cursorID),
			},
		})
	}

	limit := page.Limit
	if limit <= 0 {
		limit = 50
	}

	sql, args, err := q.OrderBy("created_at DESC", "id DESC").Limit(uint64(limit + 1)).ToSql()
	if err != nil {
		return PageResult[NotificationDBO]{}, err
	}
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return PageResult[NotificationDBO]{}, err
	}
	defer rows.Close()

	var items []NotificationDBO
	for rows.Next() {
		var n NotificationDBO
		if err := rows.Scan(&n.ID, &n.Type, &n.Title, &n.Body, &n.Read, &n.IssueID, &n.CreatedAt); err != nil {
			return PageResult[NotificationDBO]{}, err
		}
		items = append(items, n)
	}
	if err := rows.Err(); err != nil {
		return PageResult[NotificationDBO]{}, err
	}

	var nextCursor *string
	if len(items) > limit {
		items = items[:limit]
		cursor := items[len(items)-1].ID.String()
		nextCursor = &cursor
	}
	return PageResult[NotificationDBO]{Items: items, NextCursor: nextCursor}, nil
}

func (r *Repo) MarkNotificationRead(ctx context.Context, notifID, userID uuid.UUID) error {
	sql, args, err := psql.Update("notifications").
		Set("read", true).
		Where(sq.Eq{"id": notifID, "user_id": userID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) MarkAllNotificationsRead(ctx context.Context, userID uuid.UUID) error {
	sql, args, err := psql.Update("notifications").
		Set("read", true).
		Where(sq.Eq{"user_id": userID, "read": false}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

type CreateNotificationInput struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Type    string
	Title   string
	Body    *string
	IssueID *uuid.UUID
}

func (r *Repo) CreateNotification(ctx context.Context, in CreateNotificationInput) error {
	q := psql.Insert("notifications").
		Columns("id", "user_id", "type", "title", "body", "issue_id").
		Values(in.ID, in.UserID, in.Type, in.Title, in.Body, in.IssueID)
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

// ErrNoRows re-exports pgx.ErrNoRows for handlers.
var ErrNoRows = pgx.ErrNoRows
