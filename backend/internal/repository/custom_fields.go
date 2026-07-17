package repository

import (
	"context"
	"encoding/json"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) ListCustomFields(ctx context.Context, userID uuid.UUID, orgSlug, projectKey string) ([]CustomFieldDBO, error) {
	sql, args, err := psql.Select(
		"cf.id", "cf.name", "cf.key", "cf.field_type", "cf.required", "cf.options", "cf.position", "cf.created_at",
	).
		From("custom_fields cf").
		Join("projects p ON p.id = cf.project_id").
		Join("organizations o ON o.id = p.org_id").
		Join("org_members om ON om.org_id = o.id").
		Where(sq.Eq{"om.user_id": userID, "o.slug": orgSlug, "p.key": projectKey}).
		OrderBy("cf.position ASC", "cf.name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []CustomFieldDBO
	for rows.Next() {
		var f CustomFieldDBO
		if err := rows.Scan(&f.ID, &f.Name, &f.Key, &f.FieldType, &f.Required, &f.Options, &f.Position, &f.CreatedAt); err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	return fields, rows.Err()
}

func (r *Repo) CreateCustomField(ctx context.Context, id, projectID uuid.UUID, name, key, fieldType string, required bool, options []string, position int) error {
	var optionsJSON []byte
	if len(options) > 0 {
		var err error
		optionsJSON, err = json.Marshal(options)
		if err != nil {
			return err
		}
	}

	sql, args, err := psql.Insert("custom_fields").
		Columns("id", "project_id", "name", "key", "field_type", "required", "options", "position").
		Values(id, projectID, name, key, fieldType, required, optionsJSON, position).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *Repo) UpdateCustomField(ctx context.Context, fieldID, projectID uuid.UUID, name *string, required *bool, options []string, position *int) error {
	q := psql.Update("custom_fields").Where(sq.Eq{"id": fieldID, "project_id": projectID})
	if name != nil {
		q = q.Set("name", *name)
	}
	if required != nil {
		q = q.Set("required", *required)
	}
	if options != nil {
		optionsJSON, err := json.Marshal(options)
		if err != nil {
			return err
		}
		q = q.Set("options", optionsJSON)
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

func (r *Repo) DeleteCustomField(ctx context.Context, fieldID, projectID uuid.UUID) error {
	sql, args, err := psql.Delete("custom_fields").Where(sq.Eq{"id": fieldID, "project_id": projectID}).ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

type IssueCustomValueDBO struct {
	FieldID  uuid.UUID
	FieldKey string
	Value    *string
}

func (r *Repo) ListIssueCustomValues(ctx context.Context, issueID uuid.UUID) ([]IssueCustomValueDBO, error) {
	sql, args, err := psql.Select("cf.id", "cf.key", "icv.value").
		From("issue_custom_values icv").
		Join("custom_fields cf ON cf.id = icv.field_id").
		Where(sq.Eq{"icv.issue_id": issueID}).
		OrderBy("cf.position ASC", "cf.name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []IssueCustomValueDBO
	for rows.Next() {
		var v IssueCustomValueDBO
		var value *string
		if err := rows.Scan(&v.FieldID, &v.FieldKey, &value); err != nil {
			return nil, err
		}
		v.Value = value
		values = append(values, v)
	}
	return values, rows.Err()
}

func (r *Repo) UpsertIssueCustomValues(ctx context.Context, issueID uuid.UUID, values map[uuid.UUID]*string) error {
	if len(values) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for fieldID, value := range values {
		if value == nil {
			sql, args, err := psql.Delete("issue_custom_values").
				Where(sq.Eq{"issue_id": issueID, "field_id": fieldID}).
				ToSql()
			if err != nil {
				return err
			}
			if _, err = tx.Exec(ctx, sql, args...); err != nil {
				return err
			}
			continue
		}

		sql, args, err := psql.Insert("issue_custom_values").
			Columns("issue_id", "field_id", "value").
			Values(issueID, fieldID, *value).
			Suffix(`ON CONFLICT (issue_id, field_id) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`).
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

func (r *Repo) ResolveCustomFieldIDs(ctx context.Context, projectID uuid.UUID, keys []string) (map[string]uuid.UUID, error) {
	if len(keys) == 0 {
		return map[string]uuid.UUID{}, nil
	}

	sql, args, err := psql.Select("key", "id").
		From("custom_fields").
		Where(sq.Eq{"project_id": projectID, "key": keys}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]uuid.UUID, len(keys))
	for rows.Next() {
		var key string
		var id uuid.UUID
		if err := rows.Scan(&key, &id); err != nil {
			return nil, err
		}
		out[key] = id
	}
	return out, rows.Err()
}
