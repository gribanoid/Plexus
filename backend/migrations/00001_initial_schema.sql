-- +goose Up
-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- for text search

-- Enum types
CREATE TYPE global_role AS ENUM ('admin', 'user');
CREATE TYPE org_plan AS ENUM ('free', 'pro', 'enterprise');
CREATE TYPE member_role AS ENUM ('owner', 'admin', 'member', 'guest');
CREATE TYPE priority AS ENUM ('urgent', 'high', 'medium', 'low', 'no_priority');
CREATE TYPE sprint_state AS ENUM ('active', 'closed', 'future');
CREATE TYPE status_category AS ENUM ('todo', 'in_progress', 'done');
CREATE TYPE notification_type AS ENUM (
    'assigned', 'mentioned', 'commented', 'status_changed'
);

-- Users
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    display_name    TEXT NOT NULL,
    avatar_url      TEXT,
    role            global_role NOT NULL DEFAULT 'user',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);

-- Organizations (workspaces)
CREATE TABLE organizations (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    logo_url    TEXT,
    plan        org_plan NOT NULL DEFAULT 'free',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_organizations_slug ON organizations (slug);

-- Organization members
CREATE TABLE org_members (
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        member_role NOT NULL DEFAULT 'member',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (org_id, user_id)
);

-- Projects
CREATE TABLE projects (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    key         TEXT NOT NULL,
    name        TEXT NOT NULL,
    description TEXT,
    icon_url    TEXT,
    lead_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, key)
);

CREATE INDEX idx_projects_org_id ON projects (org_id);

-- Issue types (Bug, Task, Story, Epic per project)
CREATE TABLE issue_types (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    icon_url    TEXT,
    color       TEXT NOT NULL DEFAULT '#6B7280',
    UNIQUE (project_id, name)
);

-- Workflow statuses per project
CREATE TABLE statuses (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    color       TEXT NOT NULL DEFAULT '#6B7280',
    category    status_category NOT NULL DEFAULT 'todo',
    position    INT NOT NULL DEFAULT 0,
    UNIQUE (project_id, name)
);

-- Labels
CREATE TABLE labels (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    color       TEXT NOT NULL DEFAULT '#6B7280',
    UNIQUE (project_id, name)
);

-- Sprints
CREATE TABLE sprints (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    goal        TEXT,
    state       sprint_state NOT NULL DEFAULT 'future',
    start_date  TIMESTAMPTZ,
    end_date    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sprints_project_id ON sprints (project_id);

-- Issues (core entity — counter per project)
CREATE TABLE issues (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    number          BIGINT NOT NULL,
    type_id         UUID NOT NULL REFERENCES issue_types(id),
    status_id       UUID NOT NULL REFERENCES statuses(id),
    title           TEXT NOT NULL,
    description     TEXT,
    priority        priority NOT NULL DEFAULT 'no_priority',
    assignee_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    reporter_id     UUID NOT NULL REFERENCES users(id),
    parent_id       UUID REFERENCES issues(id) ON DELETE SET NULL,
    sprint_id       UUID REFERENCES sprints(id) ON DELETE SET NULL,
    story_points    REAL,
    due_date        TIMESTAMPTZ,
    position        DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, number)
);

-- Auto-increment issue number per project
CREATE SEQUENCE issue_number_seq;

CREATE INDEX idx_issues_project_id ON issues (project_id);
CREATE INDEX idx_issues_assignee_id ON issues (assignee_id);
CREATE INDEX idx_issues_status_id ON issues (status_id);
CREATE INDEX idx_issues_sprint_id ON issues (sprint_id);
CREATE INDEX idx_issues_title_trgm ON issues USING GIN (title gin_trgm_ops);

-- Issue labels (many-to-many)
CREATE TABLE issue_labels (
    issue_id    UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    label_id    UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, label_id)
);

-- Comments
CREATE TABLE comments (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    issue_id    UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author_id   UUID NOT NULL REFERENCES users(id),
    body        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_issue_id ON comments (issue_id);

-- Attachments
CREATE TABLE attachments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    issue_id        UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    uploader_id     UUID NOT NULL REFERENCES users(id),
    filename        TEXT NOT NULL,
    mime_type       TEXT NOT NULL,
    size            BIGINT NOT NULL,
    storage_key     TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_attachments_issue_id ON attachments (issue_id);

-- Issue history / audit log
CREATE TABLE issue_history (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    issue_id    UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    actor_id    UUID NOT NULL REFERENCES users(id),
    field       TEXT NOT NULL,
    old_value   TEXT,
    new_value   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_issue_history_issue_id ON issue_history (issue_id);

-- Notifications
CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        notification_type NOT NULL,
    title       TEXT NOT NULL,
    body        TEXT,
    read        BOOLEAN NOT NULL DEFAULT FALSE,
    issue_id    UUID REFERENCES issues(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications (user_id, read);

-- Refresh tokens (JWT rotation)
CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);

-- Function to auto-update updated_at
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER users_updated_at        BEFORE UPDATE ON users        FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER organizations_updated_at BEFORE UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER projects_updated_at     BEFORE UPDATE ON projects     FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER issues_updated_at       BEFORE UPDATE ON issues       FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER comments_updated_at     BEFORE UPDATE ON comments     FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS comments_updated_at ON comments;
DROP TRIGGER IF EXISTS issues_updated_at ON issues;
DROP TRIGGER IF EXISTS projects_updated_at ON projects;
DROP TRIGGER IF EXISTS organizations_updated_at ON organizations;
DROP TRIGGER IF EXISTS users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at;

DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS issue_history;
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS issue_labels;
DROP TABLE IF EXISTS issues;
DROP SEQUENCE IF EXISTS issue_number_seq;
DROP TABLE IF EXISTS sprints;
DROP TABLE IF EXISTS labels;
DROP TABLE IF EXISTS statuses;
DROP TABLE IF EXISTS issue_types;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS org_members;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS status_category;
DROP TYPE IF EXISTS sprint_state;
DROP TYPE IF EXISTS priority;
DROP TYPE IF EXISTS member_role;
DROP TYPE IF EXISTS org_plan;
DROP TYPE IF EXISTS global_role;

DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "uuid-ossp";
