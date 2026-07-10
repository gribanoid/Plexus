package repository

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) RegisterUser(ctx context.Context, in RegisterInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql, args, err := psql.Insert("users").
		Columns("id", "email", "password_hash", "display_name").
		Values(in.UserID, in.Email, in.PasswordHash, in.DisplayName).
		ToSql()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return err
	}

	sql, args, err = psql.Insert("organizations").
		Columns("id", "slug", "name").
		Values(in.OrgID, in.OrgSlug, in.OrgName).
		ToSql()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return err
	}

	sql, args, err = psql.Insert("org_members").
		Columns("org_id", "user_id", "role").
		Values(in.OrgID, in.UserID, "owner").
		ToSql()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return err
	}

	sql, args, err = psql.Insert("projects").
		Columns("id", "org_id", "key", "name", "description", "lead_id").
		Values(in.ProjectID, in.OrgID, "MAIN", "Main Project", "Your first project", in.UserID).
		ToSql()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, sql, args...); err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, "SELECT seed_project_defaults($1)", in.ProjectID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repo) GetUserCredentials(ctx context.Context, email string) (*UserCredentials, error) {
	sql, args, err := psql.Select("id", "email", "password_hash").
		From("users").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, err
	}
	var u UserCredentials
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) GetUserByRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, string, error) {
	sql, args, err := psql.Select("u.id", "u.email").
		From("refresh_tokens rt").
		Join("users u ON u.id = rt.user_id").
		Where(sq.Eq{"rt.token_hash": tokenHash}).
		Where(sq.Expr("rt.expires_at > NOW()")).
		ToSql()
	if err != nil {
		return uuid.Nil, "", err
	}
	var userID uuid.UUID
	var email string
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&userID, &email)
	return userID, email, err
}

func (r *Repo) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	sql, args, err := psql.Delete("refresh_tokens").Where(sq.Eq{"token_hash": tokenHash}).ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) InsertRefreshToken(ctx context.Context, id, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	sql, args, err := psql.Insert("refresh_tokens").
		Columns("id", "user_id", "token_hash", "expires_at").
		Values(id, userID, tokenHash, expiresAt).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfileDBO, error) {
	sql, args, err := psql.Select("id", "email", "display_name", "avatar_url", "role", "created_at").
		From("users").
		Where(sq.Eq{"id": userID}).
		ToSql()
	if err != nil {
		return nil, err
	}
	var u UserProfileDBO
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) UpdateUserProfile(ctx context.Context, userID uuid.UUID, displayName, avatarURL *string) error {
	q := psql.Update("users").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID})
	if displayName != nil {
		q = q.Set("display_name", *displayName)
	}
	if avatarURL != nil {
		q = q.Set("avatar_url", *avatarURL)
	}
	sql, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}
