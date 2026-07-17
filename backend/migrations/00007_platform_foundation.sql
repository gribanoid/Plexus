-- +goose Up

-- Backfill project members: lead + all org members as member (lead as admin)
INSERT INTO project_members (project_id, user_id, role)
SELECT p.id, p.lead_id, 'admin'::project_role
FROM projects p
WHERE p.lead_id IS NOT NULL
ON CONFLICT (project_id, user_id) DO NOTHING;

INSERT INTO project_members (project_id, user_id, role)
SELECT p.id, om.user_id, 'member'::project_role
FROM projects p
JOIN org_members om ON om.org_id = p.org_id
ON CONFLICT (project_id, user_id) DO NOTHING;

-- Workflow transitions (from_status NULL = any)
CREATE TABLE workflow_transitions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    issue_type_id   UUID REFERENCES issue_types(id) ON DELETE CASCADE,
    from_status_id  UUID REFERENCES statuses(id) ON DELETE CASCADE,
    to_status_id    UUID NOT NULL REFERENCES statuses(id) ON DELETE CASCADE,
    name            TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE NULLS NOT DISTINCT (project_id, issue_type_id, from_status_id, to_status_id)
);

CREATE INDEX idx_workflow_transitions_project ON workflow_transitions (project_id);

-- Issue links
CREATE TYPE issue_link_type AS ENUM ('blocks', 'is_blocked_by', 'relates_to', 'duplicates', 'is_duplicated_by');

CREATE TABLE issue_links (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_id    UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    target_id    UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    link_type    issue_link_type NOT NULL DEFAULT 'relates_to',
    created_by   UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source_id, target_id, link_type),
    CHECK (source_id <> target_id)
);

CREATE INDEX idx_issue_links_target ON issue_links (target_id);

-- Saved filters
CREATE TABLE saved_filters (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    owner_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    query       JSONB NOT NULL DEFAULT '{}',
    is_shared   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_saved_filters_project ON saved_filters (project_id);

CREATE TRIGGER saved_filters_updated_at
    BEFORE UPDATE ON saved_filters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Versions / releases
CREATE TABLE versions (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    description  TEXT,
    status       TEXT NOT NULL DEFAULT 'unreleased' CHECK (status IN ('unreleased', 'released', 'archived')),
    start_date   DATE,
    release_date DATE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, name)
);

CREATE TABLE issue_versions (
    issue_id    UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    version_id  UUID NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, version_id)
);

CREATE TRIGGER versions_updated_at
    BEFORE UPDATE ON versions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Components
CREATE TABLE components (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT,
    lead_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, name)
);

CREATE TABLE issue_components (
    issue_id      UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    component_id  UUID NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, component_id)
);

-- Watchers
CREATE TABLE issue_watchers (
    issue_id    UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (issue_id, user_id)
);

-- Outbound webhooks
CREATE TABLE webhooks (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    url         TEXT NOT NULL,
    secret      TEXT NOT NULL,
    events      JSONB NOT NULL DEFAULT '[]',
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_org ON webhooks (org_id);

CREATE TRIGGER webhooks_updated_at
    BEFORE UPDATE ON webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Automation rules
CREATE TABLE automation_rules (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    trigger     TEXT NOT NULL,
    conditions  JSONB NOT NULL DEFAULT '{}',
    actions     JSONB NOT NULL DEFAULT '[]',
    created_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_automation_rules_project ON automation_rules (project_id);

CREATE TRIGGER automation_rules_updated_at
    BEFORE UPDATE ON automation_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Permission schemes (enterprise scaffolding)
CREATE TABLE permission_schemes (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT,
    grants      JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, name)
);

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS permission_scheme_id UUID REFERENCES permission_schemes(id) ON DELETE SET NULL;

CREATE TRIGGER permission_schemes_updated_at
    BEFORE UPDATE ON permission_schemes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Soft-delete / trash for issues
ALTER TABLE issues
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX idx_issues_deleted_at ON issues (project_id, deleted_at) WHERE deleted_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_issues_deleted_at;
ALTER TABLE issues DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE projects DROP COLUMN IF EXISTS permission_scheme_id;
DROP TRIGGER IF EXISTS permission_schemes_updated_at ON permission_schemes;
DROP TABLE IF EXISTS permission_schemes;
DROP TRIGGER IF EXISTS automation_rules_updated_at ON automation_rules;
DROP TABLE IF EXISTS automation_rules;
DROP TRIGGER IF EXISTS webhooks_updated_at ON webhooks;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS issue_watchers;
DROP TABLE IF EXISTS issue_components;
DROP TABLE IF EXISTS components;
DROP TRIGGER IF EXISTS versions_updated_at ON versions;
DROP TABLE IF EXISTS issue_versions;
DROP TABLE IF EXISTS versions;
DROP TRIGGER IF EXISTS saved_filters_updated_at ON saved_filters;
DROP TABLE IF EXISTS saved_filters;
DROP TABLE IF EXISTS issue_links;
DROP TYPE IF EXISTS issue_link_type;
DROP TABLE IF EXISTS workflow_transitions;
