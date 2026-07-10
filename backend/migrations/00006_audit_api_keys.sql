-- +goose Up
CREATE TABLE audit_events (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id        UUID REFERENCES organizations(id) ON DELETE SET NULL,
    actor_id      UUID REFERENCES users(id) ON DELETE SET NULL,
    action        TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id   UUID,
    metadata      JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_events_org_id ON audit_events (org_id, created_at DESC);
CREATE INDEX idx_audit_events_actor_id ON audit_events (actor_id, created_at DESC);
CREATE INDEX idx_audit_events_resource ON audit_events (resource_type, resource_id);

CREATE TABLE api_keys (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id       UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    key_hash     TEXT NOT NULL UNIQUE,
    prefix       TEXT NOT NULL,
    scopes       JSONB,
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    created_by   UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at   TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_org_id ON api_keys (org_id);
CREATE INDEX idx_api_keys_prefix ON api_keys (prefix);

-- +goose Down
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS audit_events;
