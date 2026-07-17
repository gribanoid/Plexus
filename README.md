# Plexus

Enterprise project management tool — self-hosted, open source.

**Architecture:** modular Go monolith, not microservices. See [docs/architecture.md](docs/architecture.md), [docs/adr/001-modular-monolith.md](docs/adr/001-modular-monolith.md), [docs/ha-reference.md](docs/ha-reference.md), and [docs/integrations.md](docs/integrations.md).

Optional background worker: `plexus-worker` (same image; set `RUN_WORKERS=false` on API replicas).

## Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.25 + Fiber + PostgreSQL + Redis + Meilisearch |
| Web | Next.js 16 + React 19 + TypeScript + Tailwind |
| Desktop | Electron 43 + electron-vite + React (shared UI) |
| iOS | SwiftUI + Swift Concurrency |
| Android | Jetpack Compose + Kotlin + Hilt + Retrofit |
| Monorepo | Turborepo + npm workspaces |

## Layout

```
plexus/
  backend/              # Go API + worker  → backend/README.md
  web/                  # Next.js          → web/README.md
  apps/
    plexus-desktop/     # Electron         → apps/plexus-desktop/README.md
    plexus-ios/         # SwiftUI          → apps/plexus-ios/README.md
    plexus-android/     # Compose          → apps/plexus-android/README.md
  packages/
    api/                # Shared API client (web + desktop)
    ui/                 # Shared React UI
    features/           # Shared feature screens
  infra/docker/         # Local Postgres, Redis, Meilisearch, MinIO
  docs/                 # Architecture, ADR, self-hosting
```

Each runnable project has its own **README** and **Makefile** so it can later move to a separate repository with the same commands.

## Quick start

```bash
make deps          # tools check + JS/Go deps
make infra         # PostgreSQL, Redis, Meilisearch, MinIO
make migrate       # DB migrations
make seed-dev      # local admin + demo project
make dev-backend   # API on :8080
```

Then start one client:

```bash
make dev-web       # http://localhost:3000
# or
make dev-desktop
# or see mobile READMEs
```

### Dev login (`make seed-dev`)

| Field | Value |
|---|---|
| Username | `admin` (or `admin@plexus.local`) |
| Password | `admin` |

Workspace **Plexus Dev** → project **DEMO**. Do **not** run `seed-dev` in production.

## Projects

| Project | Path | Docs | Common targets |
|---|---|---|---|
| Backend | [`backend/`](backend/) | [README](backend/README.md) | `make -C backend dev` |
| Web | [`web/`](web/) | [README](web/README.md) | `make -C web dev` |
| Desktop | [`apps/plexus-desktop/`](apps/plexus-desktop/) | [README](apps/plexus-desktop/README.md) | `make -C apps/plexus-desktop dev` |
| iOS | [`apps/plexus-ios/`](apps/plexus-ios/) | [README](apps/plexus-ios/README.md) | `make -C apps/plexus-ios open` |
| Android | [`apps/plexus-android/`](apps/plexus-android/) | [README](apps/plexus-android/README.md) | `make -C apps/plexus-android run` |
| Shared JS | [`packages/`](packages/) | [README](packages/README.md) | consumed by web + desktop |

Root `make` targets (`dev-web`, `build-ios`, …) delegate into these project Makefiles.

Mobile how-to: [iOS](apps/plexus-ios/README.md) · [Android](apps/plexus-android/README.md).

## API contract

[`backend/openapi.yaml`](backend/openapi.yaml) is the HTTP contract for all clients. Web/desktop share [`@plexus/api`](packages/api); iOS/Android use hand-written clients (`APIClient.swift`, `PlexusApi.kt`).

## Self-hosting

See [docs/self-hosted.md](docs/self-hosted.md). Env reference: [`backend/.env.example`](backend/.env.example).

## Database

Migrations use [goose](https://github.com/pressly/goose) and also run on backend startup.

```bash
make migrate
make migrate-down
make seed-dev      # development only
```

Details: [backend/README.md](backend/README.md).

## Production build

```bash
make build           # backend binaries + Next.js
make build-desktop   # Electron installer
make build-ios       # Simulator (Xcode)
make build-android   # Debug APK
```

## Splitting apps (roadmap)

This monorepo is organized so clients can become separate repositories:

1. **iOS / Android** — already independent of `packages/*`; only need the API URL + OpenAPI contract.
2. **Backend + infra** — natural core repo (`openapi.yaml` remains the source of truth).
3. **Web + desktop + `packages/{api,ui,features}`** — move together, or publish `@plexus/*` and depend via registry/submodule.

Until then, develop from the repo root with `make …` as usual.
