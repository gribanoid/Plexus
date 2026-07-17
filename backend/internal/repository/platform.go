package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// --- Workflow transitions ---

type WorkflowTransitionDBO struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	IssueTypeID  sql.Null[uuid.UUID]
	FromStatusID sql.Null[uuid.UUID]
	ToStatusID   uuid.UUID
	Name         string
	CreatedAt    time.Time
}

func (r *Repo) ListWorkflowTransitions(ctx context.Context, projectID uuid.UUID) ([]WorkflowTransitionDBO, error) {
	sqlStr, args, err := psql.Select(
		"id", "project_id", "issue_type_id", "from_status_id", "to_status_id", "name", "created_at",
	).
		From("workflow_transitions").
		Where(sq.Eq{"project_id": projectID}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []WorkflowTransitionDBO
	for rows.Next() {
		var t WorkflowTransitionDBO
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.IssueTypeID, &t.FromStatusID, &t.ToStatusID, &t.Name, &t.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (r *Repo) CreateWorkflowTransition(ctx context.Context, id, projectID uuid.UUID, issueTypeID, fromStatusID *uuid.UUID, toStatusID uuid.UUID, name string) error {
	sqlStr, args, err := psql.Insert("workflow_transitions").
		Columns("id", "project_id", "issue_type_id", "from_status_id", "to_status_id", "name").
		Values(id, projectID, issueTypeID, fromStatusID, toStatusID, name).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) CountWorkflowTransitions(ctx context.Context, projectID uuid.UUID) (int, error) {
	sqlStr, args, err := psql.Select("COUNT(*)").
		From("workflow_transitions").
		Where(sq.Eq{"project_id": projectID}).
		ToSql()
	if err != nil {
		return 0, err
	}
	var n int
	err = r.pool.QueryRow(ctx, sqlStr, args...).Scan(&n)
	return n, err
}

// IsTransitionAllowed checks explicit or wildcard (NULL from_status / NULL issue_type) rules.
func (r *Repo) IsTransitionAllowed(ctx context.Context, projectID, issueTypeID, fromStatusID, toStatusID uuid.UUID) (bool, error) {
	const q = `
SELECT EXISTS (
  SELECT 1 FROM workflow_transitions
  WHERE project_id = $1
    AND to_status_id = $2
    AND (from_status_id IS NULL OR from_status_id = $3)
    AND (issue_type_id IS NULL OR issue_type_id = $4)
)`
	var ok bool
	err := r.pool.QueryRow(ctx, q, projectID, toStatusID, fromStatusID, issueTypeID).Scan(&ok)
	return ok, err
}

func (r *Repo) DeleteWorkflowTransition(ctx context.Context, id, projectID uuid.UUID) error {
	sqlStr, args, err := psql.Delete("workflow_transitions").
		Where(sq.Eq{"id": id, "project_id": projectID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

// --- Issue links ---

type IssueLinkDBO struct {
	ID        uuid.UUID
	SourceID  uuid.UUID
	TargetID  uuid.UUID
	LinkType  string
	CreatedBy sql.Null[uuid.UUID]
	CreatedAt time.Time
}

func (r *Repo) ListIssueLinks(ctx context.Context, issueID uuid.UUID) ([]IssueLinkDBO, error) {
	sqlStr, args, err := psql.Select("id", "source_id", "target_id", "link_type", "created_by", "created_at").
		From("issue_links").
		Where(sq.Or{sq.Eq{"source_id": issueID}, sq.Eq{"target_id": issueID}}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []IssueLinkDBO
	for rows.Next() {
		var l IssueLinkDBO
		if err := rows.Scan(&l.ID, &l.SourceID, &l.TargetID, &l.LinkType, &l.CreatedBy, &l.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, l)
	}
	return items, rows.Err()
}

func (r *Repo) CreateIssueLink(ctx context.Context, id, sourceID, targetID uuid.UUID, linkType string, createdBy *uuid.UUID) error {
	sqlStr, args, err := psql.Insert("issue_links").
		Columns("id", "source_id", "target_id", "link_type", "created_by").
		Values(id, sourceID, targetID, linkType, createdBy).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) DeleteIssueLink(ctx context.Context, id, projectID uuid.UUID) error {
	sqlStr, args, err := psql.Delete("issue_links").
		Where(sq.Eq{"id": id}).
		Where(sq.Expr(`EXISTS (
			SELECT 1 FROM issues i
			WHERE i.project_id = ? AND (i.id = issue_links.source_id OR i.id = issue_links.target_id)
		)`, projectID)).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

// --- Saved filters ---

type SavedFilterDBO struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	OwnerID   uuid.UUID
	Name      string
	Query     []byte
	IsShared  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *Repo) ListSavedFilters(ctx context.Context, projectID, userID uuid.UUID) ([]SavedFilterDBO, error) {
	sqlStr, args, err := psql.Select(
		"id", "project_id", "owner_id", "name", "query", "is_shared", "created_at", "updated_at",
	).
		From("saved_filters").
		Where(sq.Eq{"project_id": projectID}).
		Where(sq.Or{sq.Eq{"owner_id": userID}, sq.Eq{"is_shared": true}}).
		OrderBy("name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SavedFilterDBO
	for rows.Next() {
		var f SavedFilterDBO
		if err := rows.Scan(&f.ID, &f.ProjectID, &f.OwnerID, &f.Name, &f.Query, &f.IsShared, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, f)
	}
	return items, rows.Err()
}

func (r *Repo) CreateSavedFilter(ctx context.Context, id, projectID, ownerID uuid.UUID, name string, query json.RawMessage, isShared bool) error {
	if query == nil {
		query = json.RawMessage("{}")
	}
	sqlStr, args, err := psql.Insert("saved_filters").
		Columns("id", "project_id", "owner_id", "name", "query", "is_shared").
		Values(id, projectID, ownerID, name, []byte(query), isShared).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) UpdateSavedFilter(ctx context.Context, id, ownerID uuid.UUID, name *string, query json.RawMessage, isShared *bool) error {
	q := psql.Update("saved_filters").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id, "owner_id": ownerID})
	if name != nil {
		q = q.Set("name", *name)
	}
	if query != nil {
		q = q.Set("query", []byte(query))
	}
	if isShared != nil {
		q = q.Set("is_shared", *isShared)
	}
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) DeleteSavedFilter(ctx context.Context, id, ownerID uuid.UUID) error {
	sqlStr, args, err := psql.Delete("saved_filters").
		Where(sq.Eq{"id": id, "owner_id": ownerID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

// --- Versions ---

type VersionDBO struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Name        string
	Description sql.NullString
	Status      string
	StartDate   sql.NullTime
	ReleaseDate sql.NullTime
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r *Repo) ListVersions(ctx context.Context, projectID uuid.UUID) ([]VersionDBO, error) {
	sqlStr, args, err := psql.Select(
		"id", "project_id", "name", "description", "status", "start_date", "release_date", "created_at", "updated_at",
	).
		From("versions").
		Where(sq.Eq{"project_id": projectID}).
		OrderBy("name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []VersionDBO
	for rows.Next() {
		var v VersionDBO
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.Name, &v.Description, &v.Status, &v.StartDate, &v.ReleaseDate, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, v)
	}
	return items, rows.Err()
}

func (r *Repo) CreateVersion(ctx context.Context, id, projectID uuid.UUID, name string, description *string, status string, startDate, releaseDate *time.Time) error {
	if status == "" {
		status = "unreleased"
	}
	sqlStr, args, err := psql.Insert("versions").
		Columns("id", "project_id", "name", "description", "status", "start_date", "release_date").
		Values(id, projectID, name, description, status, startDate, releaseDate).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) UpdateVersion(ctx context.Context, id, projectID uuid.UUID, name, description, status *string, startDate, releaseDate *time.Time) error {
	q := psql.Update("versions").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id, "project_id": projectID})
	if name != nil {
		q = q.Set("name", *name)
	}
	if description != nil {
		q = q.Set("description", *description)
	}
	if status != nil {
		q = q.Set("status", *status)
	}
	if startDate != nil {
		q = q.Set("start_date", *startDate)
	}
	if releaseDate != nil {
		q = q.Set("release_date", *releaseDate)
	}
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) DeleteVersion(ctx context.Context, id, projectID uuid.UUID) error {
	sqlStr, args, err := psql.Delete("versions").Where(sq.Eq{"id": id, "project_id": projectID}).ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) SetIssueVersions(ctx context.Context, issueID uuid.UUID, versionIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	delSQL, delArgs, err := psql.Delete("issue_versions").Where(sq.Eq{"issue_id": issueID}).ToSql()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, delSQL, delArgs...); err != nil {
		return err
	}
	for _, vid := range versionIDs {
		insSQL, insArgs, err := psql.Insert("issue_versions").
			Columns("issue_id", "version_id").
			Values(issueID, vid).
			ToSql()
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, insSQL, insArgs...); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// --- Components ---

type ComponentDBO struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Name        string
	Description sql.NullString
	LeadID      sql.Null[uuid.UUID]
	CreatedAt   time.Time
}

func (r *Repo) ListComponents(ctx context.Context, projectID uuid.UUID) ([]ComponentDBO, error) {
	sqlStr, args, err := psql.Select("id", "project_id", "name", "description", "lead_id", "created_at").
		From("components").
		Where(sq.Eq{"project_id": projectID}).
		OrderBy("name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ComponentDBO
	for rows.Next() {
		var c ComponentDBO
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Name, &c.Description, &c.LeadID, &c.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *Repo) CreateComponent(ctx context.Context, id, projectID uuid.UUID, name string, description *string, leadID *uuid.UUID) error {
	sqlStr, args, err := psql.Insert("components").
		Columns("id", "project_id", "name", "description", "lead_id").
		Values(id, projectID, name, description, leadID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) DeleteComponent(ctx context.Context, id, projectID uuid.UUID) error {
	sqlStr, args, err := psql.Delete("components").Where(sq.Eq{"id": id, "project_id": projectID}).ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) SetIssueComponents(ctx context.Context, issueID uuid.UUID, componentIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	delSQL, delArgs, err := psql.Delete("issue_components").Where(sq.Eq{"issue_id": issueID}).ToSql()
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, delSQL, delArgs...); err != nil {
		return err
	}
	for _, cid := range componentIDs {
		insSQL, insArgs, err := psql.Insert("issue_components").
			Columns("issue_id", "component_id").
			Values(issueID, cid).
			ToSql()
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, insSQL, insArgs...); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// --- Watchers ---

type WatcherDBO struct {
	UserID      uuid.UUID
	DisplayName string
	Email       string
	AvatarURL   sql.NullString
	CreatedAt   time.Time
}

func (r *Repo) ListWatchers(ctx context.Context, issueID uuid.UUID) ([]WatcherDBO, error) {
	sqlStr, args, err := psql.Select("u.id", "u.display_name", "u.email", "u.avatar_url", "iw.created_at").
		From("issue_watchers iw").
		Join("users u ON u.id = iw.user_id").
		Where(sq.Eq{"iw.issue_id": issueID}).
		OrderBy("u.display_name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []WatcherDBO
	for rows.Next() {
		var w WatcherDBO
		if err := rows.Scan(&w.UserID, &w.DisplayName, &w.Email, &w.AvatarURL, &w.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, rows.Err()
}

func (r *Repo) AddWatcher(ctx context.Context, issueID, userID uuid.UUID) error {
	sqlStr, args, err := psql.Insert("issue_watchers").
		Columns("issue_id", "user_id").
		Values(issueID, userID).
		Suffix("ON CONFLICT DO NOTHING").
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) RemoveWatcher(ctx context.Context, issueID, userID uuid.UUID) error {
	sqlStr, args, err := psql.Delete("issue_watchers").
		Where(sq.Eq{"issue_id": issueID, "user_id": userID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

// --- Webhooks ---

type WebhookDBO struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	Name      string
	URL       string
	Secret    string
	Events    []byte
	Active    bool
	CreatedBy sql.Null[uuid.UUID]
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *Repo) ListWebhooks(ctx context.Context, orgID uuid.UUID) ([]WebhookDBO, error) {
	sqlStr, args, err := psql.Select(
		"id", "org_id", "name", "url", "secret", "events", "active", "created_by", "created_at", "updated_at",
	).
		From("webhooks").
		Where(sq.Eq{"org_id": orgID}).
		OrderBy("name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []WebhookDBO
	for rows.Next() {
		var w WebhookDBO
		if err := rows.Scan(&w.ID, &w.OrgID, &w.Name, &w.URL, &w.Secret, &w.Events, &w.Active, &w.CreatedBy, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, rows.Err()
}

func (r *Repo) CreateWebhook(ctx context.Context, id, orgID uuid.UUID, name, url, secret string, events []string, createdBy *uuid.UUID) error {
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return err
	}
	if eventsJSON == nil {
		eventsJSON = []byte("[]")
	}
	sqlStr, args, err := psql.Insert("webhooks").
		Columns("id", "org_id", "name", "url", "secret", "events", "created_by").
		Values(id, orgID, name, url, secret, eventsJSON, createdBy).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) UpdateWebhook(ctx context.Context, id, orgID uuid.UUID, name, url, secret *string, events []string, active *bool) error {
	q := psql.Update("webhooks").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id, "org_id": orgID})
	if name != nil {
		q = q.Set("name", *name)
	}
	if url != nil {
		q = q.Set("url", *url)
	}
	if secret != nil {
		q = q.Set("secret", *secret)
	}
	if events != nil {
		eventsJSON, err := json.Marshal(events)
		if err != nil {
			return err
		}
		q = q.Set("events", eventsJSON)
	}
	if active != nil {
		q = q.Set("active", *active)
	}
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) DeleteWebhook(ctx context.Context, id, orgID uuid.UUID) error {
	sqlStr, args, err := psql.Delete("webhooks").
		Where(sq.Eq{"id": id, "org_id": orgID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

// --- Automation rules ---

type AutomationRuleDBO struct {
	ID         uuid.UUID
	ProjectID  uuid.UUID
	Name       string
	Enabled    bool
	Trigger    string
	Conditions []byte
	Actions    []byte
	CreatedBy  sql.Null[uuid.UUID]
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (r *Repo) ListAutomationRules(ctx context.Context, projectID uuid.UUID) ([]AutomationRuleDBO, error) {
	sqlStr, args, err := psql.Select(
		"id", "project_id", "name", "enabled", "trigger", "conditions", "actions", "created_by", "created_at", "updated_at",
	).
		From("automation_rules").
		Where(sq.Eq{"project_id": projectID}).
		OrderBy("name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AutomationRuleDBO
	for rows.Next() {
		var a AutomationRuleDBO
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.Name, &a.Enabled, &a.Trigger, &a.Conditions, &a.Actions, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

func (r *Repo) CreateAutomationRule(ctx context.Context, id, projectID uuid.UUID, name, trigger string, conditions, actions json.RawMessage, createdBy *uuid.UUID) error {
	if conditions == nil {
		conditions = json.RawMessage("{}")
	}
	if actions == nil {
		actions = json.RawMessage("[]")
	}
	sqlStr, args, err := psql.Insert("automation_rules").
		Columns("id", "project_id", "name", "trigger", "conditions", "actions", "created_by").
		Values(id, projectID, name, trigger, []byte(conditions), []byte(actions), createdBy).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) UpdateAutomationRule(ctx context.Context, id, projectID uuid.UUID, name, trigger *string, enabled *bool, conditions, actions json.RawMessage) error {
	q := psql.Update("automation_rules").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id, "project_id": projectID})
	if name != nil {
		q = q.Set("name", *name)
	}
	if trigger != nil {
		q = q.Set("trigger", *trigger)
	}
	if enabled != nil {
		q = q.Set("enabled", *enabled)
	}
	if conditions != nil {
		q = q.Set("conditions", []byte(conditions))
	}
	if actions != nil {
		q = q.Set("actions", []byte(actions))
	}
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) DeleteAutomationRule(ctx context.Context, id, projectID uuid.UUID) error {
	sqlStr, args, err := psql.Delete("automation_rules").
		Where(sq.Eq{"id": id, "project_id": projectID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

// --- Permission schemes ---

type PermissionSchemeDBO struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	Name        string
	Description sql.NullString
	Grants      []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r *Repo) ListPermissionSchemes(ctx context.Context, orgID uuid.UUID) ([]PermissionSchemeDBO, error) {
	sqlStr, args, err := psql.Select(
		"id", "org_id", "name", "description", "grants", "created_at", "updated_at",
	).
		From("permission_schemes").
		Where(sq.Eq{"org_id": orgID}).
		OrderBy("name ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PermissionSchemeDBO
	for rows.Next() {
		var p PermissionSchemeDBO
		if err := rows.Scan(&p.ID, &p.OrgID, &p.Name, &p.Description, &p.Grants, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *Repo) CreatePermissionScheme(ctx context.Context, id, orgID uuid.UUID, name string, description *string, grants json.RawMessage) error {
	if grants == nil {
		grants = json.RawMessage("{}")
	}
	sqlStr, args, err := psql.Insert("permission_schemes").
		Columns("id", "org_id", "name", "description", "grants").
		Values(id, orgID, name, description, []byte(grants)).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) UpdatePermissionScheme(ctx context.Context, id, orgID uuid.UUID, name, description *string, grants json.RawMessage) error {
	q := psql.Update("permission_schemes").
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id, "org_id": orgID})
	if name != nil {
		q = q.Set("name", *name)
	}
	if description != nil {
		q = q.Set("description", *description)
	}
	if grants != nil {
		q = q.Set("grants", []byte(grants))
	}
	sqlStr, args, err := q.ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) AssignPermissionScheme(ctx context.Context, projectID uuid.UUID, schemeID *uuid.UUID) error {
	sqlStr, args, err := psql.Update("projects").
		Set("permission_scheme_id", schemeID).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": projectID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

// --- Soft-delete ---

func (r *Repo) SetIssueDeletedAt(ctx context.Context, issueID uuid.UUID, deletedAt *time.Time) error {
	sqlStr, args, err := psql.Update("issues").
		Set("deleted_at", deletedAt).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": issueID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, sqlStr, args...)
	return err
}

func (r *Repo) GetIssueIDIncludingDeleted(ctx context.Context, projectID uuid.UUID, issueNumber int64) (uuid.UUID, error) {
	sqlStr, args, err := psql.Select("id").
		From("issues").
		Where(sq.Eq{"project_id": projectID, "number": issueNumber}).
		ToSql()
	if err != nil {
		return uuid.Nil, err
	}
	var id uuid.UUID
	err = r.pool.QueryRow(ctx, sqlStr, args...).Scan(&id)
	return id, err
}

func (r *Repo) GetWebhookByID(ctx context.Context, id uuid.UUID) (*WebhookDBO, error) {
	sqlStr, args, err := psql.Select(
		"id", "org_id", "name", "url", "secret", "events", "active", "created_by", "created_at", "updated_at",
	).
		From("webhooks").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}
	var w WebhookDBO
	err = r.pool.QueryRow(ctx, sqlStr, args...).Scan(
		&w.ID, &w.OrgID, &w.Name, &w.URL, &w.Secret, &w.Events, &w.Active, &w.CreatedBy, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// ListActiveWebhooksForEvent returns org webhooks subscribed to event (or "*" / empty = all).
func (r *Repo) ListActiveWebhooksForEvent(ctx context.Context, orgID uuid.UUID, event string) ([]WebhookDBO, error) {
	all, err := r.ListWebhooks(ctx, orgID)
	if err != nil {
		return nil, err
	}
	var out []WebhookDBO
	for _, wh := range all {
		if !wh.Active {
			continue
		}
		var events []string
		_ = json.Unmarshal(wh.Events, &events)
		if len(events) == 0 {
			out = append(out, wh)
			continue
		}
		for _, e := range events {
			if e == "*" || e == event {
				out = append(out, wh)
				break
			}
		}
	}
	return out, nil
}

func (r *Repo) ListEnabledAutomationRules(ctx context.Context, projectID uuid.UUID, trigger string) ([]AutomationRuleDBO, error) {
	all, err := r.ListAutomationRules(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var out []AutomationRuleDBO
	for _, rule := range all {
		if rule.Enabled && rule.Trigger == trigger {
			out = append(out, rule)
		}
	}
	return out, nil
}

