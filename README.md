# Plexus

Enterprise project management tool — self-hosted, open source.

## Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.25 + Fiber + PostgreSQL + Redis + Meilisearch |
| Web | Next.js 15 + React 19 + TypeScript + Tailwind + shadcn/ui |
| Desktop | Electron 43 + electron-vite + React (shared UI) |
| iOS | SwiftUI + Swift Concurrency |
| Android | Jetpack Compose + Kotlin + Hilt + Retrofit |
| Monorepo | Turborepo + npm workspaces |

## Project structure

```
plexus/
  backend/           # Go API server
    cmd/server/      # Entrypoint
    internal/        # Business logic
    migrations/      # SQL migrations (goose)
    openapi.yaml     # API contract (source of truth for all clients)
  apps/
    web/             # Next.js web application
    desktop/         # Electron desktop application
    ios/             # SwiftUI iOS application
    android/         # Jetpack Compose Android application
  packages/
    ui/              # Shared React components (web + desktop)
    api/             # Shared API layer for web + desktop (Phase 2)
  infra/
    docker/          # Docker Compose for local dev
  scripts/
    generate-api-clients.sh   # Regenerate Swift & Kotlin clients from openapi.yaml
```

## Quick start (local development)

### Prerequisites

```bash
make deps          # verify tools + install JS/Go dependencies
make infra         # PostgreSQL, Redis, Meilisearch, MinIO
make migrate       # apply database migrations
make seed-dev      # dev-only: admin user + demo project (see below)
```

### 1. Start infrastructure services

```bash
make infra
# or: cd infra/docker && docker compose up -d postgres redis meilisearch minio
```

### 2. Start the Go backend

```bash
make dev-backend
# or: cd backend && cp .env.example .env && go run ./cmd/server
```

### Dev login (`make seed-dev`)

After `make infra`, `make migrate`, and `make seed-dev`:

| Field | Value |
|---|---|
| Username | `admin` (or `admin@plexus.local`) |
| Password | `admin` |

Opens workspace **Plexus Dev** → project **DEMO** with sample issues.

> Dev seed is **not** applied in production. Use `make seed-dev` only in local development.

### 3. Start the web app

```bash
make dev-web        # → http://localhost:3000
```

If port 3000 is already in use:

```bash
make dev-web-stop   # kill the process on port 3000
make dev-web
```

### 4. Start Electron desktop

```bash
make dev-desktop
```

## Mobile builds

| Platform | Requirements | Command |
|---|---|---|
| iOS | macOS, Xcode 15+ | `make deps-ios && make build-ios` |
| Android | Java 17+, Android SDK | `make deps-android && make build-android` |

Run `make deps-check` to see which optional tools are missing.

## Self-hosting

See [docs/self-hosted.md](docs/self-hosted.md) for a production deployment overview.

## Generating API clients

After changing `backend/openapi.yaml`, regenerate all clients:

```bash
# Requires: brew install openapi-generator  (or npm install -g @openapitools/openapi-generator-cli)
./scripts/generate-api-clients.sh
```

## Database migrations

Migrations use [goose](https://github.com/pressly/goose). They run automatically when the backend starts.

```bash
# Apply all pending migrations manually
make migrate

# Roll back the last migration
make migrate-down

# Or start the backend (migrations run on startup)
make dev-backend
```

## Environment variables

See `backend/.env.example` for all available configuration.

## Building for production

```bash
make build           # backend binary + Next.js
make build-desktop   # Electron installer
```
