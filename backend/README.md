# Plexus backend

Go API server and optional background worker for Plexus.

**Stack:** Go 1.25 · Fiber · PostgreSQL · Redis · Meilisearch · MinIO/S3

HTTP contract: [`openapi.yaml`](openapi.yaml) (source of truth for all clients).

## Prerequisites

- Go 1.22+ (repo uses 1.25 toolchain)
- Docker (for local Postgres / Redis / Meilisearch / MinIO)
- `psql` (for `make seed-dev`)

## Setup

```bash
cp .env.example .env   # or: make -C backend will create it on first `dev`
make -C backend deps
make -C backend infra  # from monorepo: uses ../infra/docker
make -C backend migrate
make -C backend seed-dev   # development only
make -C backend dev        # :8080
```

From the monorepo root the same flow is `make deps && make infra && make migrate && make seed-dev && make dev-backend`.

## Commands

| Target | Description |
|---|---|
| `make deps` | `go mod download` |
| `make infra` / `infra-down` | Start/stop local infra via Compose |
| `make dev` | Run `cmd/server` (migrations on startup) |
| `make build` | Binaries → `dist/plexus-server` + `dist/plexus-worker` |
| `make migrate` / `migrate-down` | Goose migrations |
| `make seed-dev` | Admin + DEMO project (**not for production**) |
| `make test` | `go test ./...` |
| `make fmt` | `gofmt` |

## Configuration

See [`.env.example`](.env.example). Important for clients:

| Variable | Role |
|---|---|
| `CORS_ORIGINS` | Allowed browser/Electron origins (include web `:3000` and desktop Vite `:5173`) |
| `FRONTEND_URL` | OIDC / redirects |
| `JWT_SECRET` | Must be strong in production |
| `RUN_WORKERS` | `false` on API replicas when running dedicated `plexus-worker` |

Default local CORS includes `http://localhost:3000`, `http://localhost:5173`, `http://127.0.0.1:5173`, `app://plexus`.

## Dev login (after `seed-dev`)

| Field | Value |
|---|---|
| Email / username | `admin@plexus.local` or `admin` |
| Password | `admin` |

## Clients

| Client | Default API base |
|---|---|
| Web | same-origin `/api/v1` (Next rewrite → `:8080`) |
| Desktop | `http://127.0.0.1:8080/api/v1` (`VITE_API_URL`) |
| iOS Simulator | `http://127.0.0.1:8080/api/v1` |
| Android emulator | `http://10.0.2.2:8080/api/v1` |

## Future: standalone repo

Move with `migrations/`, `openapi.yaml`, `scripts/`, and preferably `infra/docker/` (or document external Compose). Clients become separate repos that depend only on this API + OpenAPI. Extend `CORS_ORIGINS` / `FRONTEND_URL` for each deployed client origin.
