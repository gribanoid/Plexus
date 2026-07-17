# Self-hosting Plexus

Plexus is designed to run on your own infrastructure. This guide covers a minimal production deployment.

Architecture: **modular monolith** — see [architecture.md](architecture.md), [ADR 001](adr/001-modular-monolith.md), and [ha-reference.md](ha-reference.md) for HA sizing. Integrations: [integrations.md](integrations.md).

## Requirements

- Docker (or PostgreSQL 16+, Redis 7+, Meilisearch 1.x, S3-compatible storage)
- Go 1.25+ (to build the API server) or pre-built `plexus-server` / `plexus-worker` binaries
- Node.js 20+ (to build the web app) or a pre-built Next.js output

## Quick deploy with Docker Compose

1. Copy `backend/.env.example` to `backend/.env` and set production values:
   - `DATABASE_URL`, `REDIS_URL`, `JWT_SECRET` (long random string)
   - `MEILISEARCH_URL`, `MEILISEARCH_API_KEY`
   - `S3_*` variables for attachments
   - `FRONTEND_URL` and `CORS_ORIGINS` (your web app URL)
   - Optional: `RUN_WORKERS=false` on API replicas when running dedicated `plexus-worker`

2. Start infrastructure:

```bash
make infra
make migrate
```

3. Build and run:

```bash
make build-backend
make build-web
./dist/plexus-server   # API on :8080 (embeds workers by default)
# Optional dedicated worker:
# RUN_WORKERS=false ./dist/plexus-server &
# ./dist/plexus-worker
npm run start --workspace=web   # Web on :3000
# or: make -C web start
```

4. Create the first user via `POST /api/v1/auth/register` (requires `ALLOW_REGISTRATION=true` outside development), seed-dev in non-prod only, or enable OIDC in `.env`. Set `METRICS_AUTH_TOKEN` for production.

> Do **not** run `make seed-dev` in production — it creates a known admin account.

## Observability

API exposes:

| Endpoint | Port | Auth |
|----------|------|------|
| `GET /health` | `PORT` (8080) | none (liveness) |
| `GET /health/detailed` | 8080 | `METRICS_AUTH_TOKEN` |
| `GET /metrics` | `METRICS_LISTEN_ADDRESS` (default `:9090`) | Bearer / `X-Metrics-Token` |
| `/debug/pprof/*` | `:9090` | same token |

Example profile:

```bash
go tool pprof http://localhost:9090/debug/pprof/profile?seconds=30
# with token:
curl -H "Authorization: Bearer $METRICS_AUTH_TOKEN" http://localhost:9090/metrics
```

Local Prometheus + Grafana:

```bash
cd infra/docker
docker compose --profile monitoring up -d
# Prometheus UI: http://localhost:9091
# Grafana: http://localhost:3001 (admin/admin) — dashboard "Plexus RED"
```

Env: `METRICS_ENABLED`, `METRICS_LISTEN_ADDRESS`, `METRICS_AUTH_TOKEN`, `LOG_LEVEL`.

## Reverse proxy

Place nginx or Caddy in front of both services:

| Path | Backend |
|---|---|
| `/api/v1/*` | Go API (`:8080`) |
| `/ws` | WebSocket upgrade to API |
| `/*` | Next.js (`:3000`) |

Do **not** expose `:9090` (metrics/pprof) publicly; scrape from a private network or via VPN. Enable TLS and set `FRONTEND_URL` to your public HTTPS origin.

## Health checks

- `GET /health` — liveness
- `GET /health/detailed` — PostgreSQL and Redis status (requires metrics token outside open local auth)

## Backups

Back up PostgreSQL regularly. Redis holds ephemeral data (WebSocket fan-out, job queue, rate limits) and does not require backup for correctness.

## Updates

1. Pull the new release
2. `make migrate`
3. Rebuild and restart backend + web
4. Reindex search if schema changed: issues are re-indexed automatically via background jobs
