-- +goose Up
CREATE TYPE custom_field_type AS ENUM ('text', 'number', 'date', 'select', 'boolean');

CREATE TABLE custom_fields (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    key         TEXT NOT NULL,
    field_type  custom_field_type NOT NULL DEFAULT 'text',
    required    BOOLEAN NOT NULL DEFAULT FALSE,
    options     JSONB,
    position    INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, key)
);

CREATE INDEX idx_custom_fields_project_id ON custom_fields (project_id);

CREATE TABLE issue_custom_values (
    issue_id    UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    field_id    UUID NOT NULL REFERENCES custom_fields(id) ON DELETE CASCADE,
    value       TEXT,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (issue_id, field_id)
);

CREATE INDEX idx_issue_custom_values_field_id ON issue_custom_values (field_id);

CREATE TRIGGER custom_fields_updated_at
    BEFORE UPDATE ON custom_fields
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS custom_fields_updated_at ON custom_fields;
DROP TABLE IF EXISTS issue_custom_values;
DROP TABLE IF EXISTS custom_fields;
DROP TYPE IF EXISTS custom_field_type;
