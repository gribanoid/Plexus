# Plexus Web

Next.js web client for Plexus.

**Stack:** Next.js 16 · React 19 · TypeScript · Tailwind · TanStack Query · shared `@plexus/{api,ui,features}`

## Prerequisites

- Node.js ≥ 20, npm ≥ 10
- Running [backend](../backend/README.md) on `:8080` (plus infra)

## Setup

From monorepo root:

```bash
make deps
make infra && make migrate && make seed-dev
make dev-backend   # other terminal
make -C web dev
# → http://localhost:3000
```

Or: `make dev-web` from the repo root.

## Commands

| Target | Description |
|---|---|
| `make deps` | Install monorepo JS dependencies |
| `make dev` | Dev server on `:3000` |
| `make stop` | Kill process on `:3000` |
| `make build` / `start` | Production build & serve |
| `make typecheck` / `lint` | Quality |
| `make test-e2e` | Playwright |

## API URL

Browser requests use same-origin `/api/v1`; Next.js rewrites to the backend (`NEXT_PUBLIC_API_URL`, default `http://localhost:8080`).

SSR and server code use `${NEXT_PUBLIC_API_URL}/api/v1`.

Ensure backend `CORS_ORIGINS` includes your web origin if you call the API cross-origin without the rewrite.

## Shared packages

This app depends on:

- `@plexus/api` — auth, fetch, hooks
- `@plexus/ui` — UI primitives
- `@plexus/features` — issue/project dialogs, search, notifications

See [packages/README.md](../packages/README.md).

## Dev login

After backend `seed-dev`: `admin` / `admin`.

## Future: standalone repo

Ship together with `packages/{api,ui,features}` (or publish `@plexus/*` to a registry). Point `NEXT_PUBLIC_API_URL` at the API host; keep OpenAPI contract from the backend repo in sync.
