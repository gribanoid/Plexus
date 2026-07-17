# High availability reference architectures

Plexus follows the Mattermost scaling model: **replicate the monolith**, share data stores. Kubernetes is optional, not required.

## Small (≤ 50 concurrent users)

| Component | Recommendation |
|-----------|----------------|
| API | 1× `plexus-server` (embeds workers) |
| Postgres | 1 instance (managed or Compose) |
| Redis | 1 instance |
| Meilisearch | 1 instance |
| Object storage | MinIO or S3 |
| Proxy | Caddy/nginx TLS termination |

Compose in `infra/docker/` is sufficient.

## Medium (50–500 concurrent users)

| Component | Recommendation |
|-----------|----------------|
| API | 2–3× `plexus-server` behind LB |
| Worker | 1–2× `plexus-worker` (`PLEXUS_RUN_WORKERS=false` on API) |
| Postgres | Primary with automated backups; consider read replica for heavy search/report reads |
| Redis | Single or managed Redis with persistence |
| Meilisearch | Dedicated node with disk sized for issue corpus |
| Proxy | Sticky sessions optional for WS; Redis pub/sub already fans out events |

## Large / HA (500+ concurrent, enterprise)

| Component | Recommendation |
|-----------|----------------|
| API | ≥ 3 replicas, multi-AZ |
| Worker | ≥ 2 replicas |
| Postgres | Primary + standby (failover); connection pooling (PgBouncer) |
| Redis | Sentinel/Cluster or managed HA |
| Search | Meilisearch HA or dedicated search tier |
| Metrics | Prometheus scrape `/metrics` (when enabled) |
| Backups | Daily DB + object storage versioning; tested restore runbook |

### WebSocket notes

Clients connect to `/ws?token=…&project_id=…`. With multiple API nodes, events are published on Redis channel `plexus:events` so every node fans out to local connections. Prefer proxy support for long-lived WebSocket upgrades.

### Compliance controls (product + infra)

| Control | Where |
|---------|--------|
| Audit log | DB `audit_events` + org admin list API |
| SSO | OIDC (shipped); SAML/LDAP via identity broker or future native |
| Encryption in transit | TLS at reverse proxy |
| Encryption at rest | Cloud disk / Postgres / S3 SSE (infra) |
| API keys | Org-scoped keys with hash storage + revoke |
| Soft-delete / retention | Trash + retention policies (roadmap); export via CSV |

## Non-goals

- Domain microservices and service mesh as a default.
- Separate database per bounded context.
