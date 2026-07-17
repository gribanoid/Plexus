# Plexus architecture

## Overview

Plexus is a **modular monolith**: one Go backend process serves REST + WebSocket; optional worker process runs background jobs. Clients (web, desktop, iOS, Android) talk to `/api/v1` and `/ws`.

See [ADR 001](adr/001-modular-monolith.md) for the microservices decision.

```
Access          Application                         Infra
─────           ───────────                         ─────
Web/Desktop  →  Fiber REST (/api/v1)            →   PostgreSQL
Mobile       →  Domain services (in-process)    →   Redis (cache, pub/sub, asynq)
             →  WebSocket hub                   →   Meilisearch
             →  asynq workers (API or worker)   →   S3/MinIO
```

## Backend layout

| Path | Role |
|------|------|
| `cmd/server` | HTTP API + WS + (by default) embedded workers |
| `cmd/worker` | Background jobs only (`RUN_MODE=worker` / dedicated binary) |
| `internal/api` | Fiber routes, thin handlers |
| `internal/service` | Business logic (authz, workflows, notifications, …) |
| `internal/repository` | SQL (Squirrel + pgx) |
| `internal/jobs` | asynq task handlers |
| `internal/websocket` | Realtime fan-out via Redis |
| `internal/search` | Meilisearch client |
| `migrations/` | goose SQL |

## AuthZ model

1. **Org roles**: `owner` | `admin` | `member` | `guest` — membership gate + org admin ops.
2. **Project roles**: `admin` | `member` | `viewer` — enforced on project-scoped routes.
3. Org `owner`/`admin` bypass project membership (full project access).
4. Guests are read-only at org level; project `viewer` cannot write.

## Scaling path (Mattermost-like)

1. Single node: API embeds workers.
2. Split: `plexus-server` (API/WS) + `plexus-worker` (jobs).
3. HA: N API replicas behind reverse proxy; Redis for WS pub/sub and queues; Postgres primary (+ replicas later).

Details: [ha-reference.md](ha-reference.md).

## Extensibility

Outbound webhooks and automation rules run inside the monolith (and worker). Marketplace-style integrations consume the public OpenAPI and signed webhook events — not separate core services.
