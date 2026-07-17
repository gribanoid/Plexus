# Plexus Desktop

Electron desktop client for Plexus (shared React UI with web).

**Stack:** Electron 43 · electron-vite · React 19 · React Router · `@plexus/{api,ui,features}`

## Prerequisites

- Node.js ≥ 20, npm ≥ 10
- Running [backend](../../backend/README.md) on `:8080`
- Electron binary installed (`node node_modules/electron/install.js` if skipped)

## Setup

```bash
make deps
make infra && make migrate && make seed-dev
make dev-backend
make -C apps/plexus-desktop dev
```

Or from root: `make dev-desktop`.

## Commands

| Target | Description |
|---|---|
| `make deps` | Install monorepo JS deps |
| `make dev` | Electron + Vite (`localhost:5173` renderer) |
| `make build` | Production renderer/main build |
| `make preview` | Run built app |
| `make package` | OS installers via electron-builder |
| `make typecheck` / `lint` | Quality |

## API URL

Copy [`.env.example`](.env.example) → `.env` if needed:

```
VITE_API_URL=http://127.0.0.1:8080/api/v1
```

Backend `CORS_ORIGINS` must allow the Vite origin (`http://localhost:5173` / `http://127.0.0.1:5173`).

**Note:** IDEs like Cursor may set `ELECTRON_RUN_AS_NODE=1`, which breaks Electron. The Makefile clears it (`env -u ELECTRON_RUN_AS_NODE`).

## Shared packages

Same as web: `@plexus/api`, `@plexus/ui`, `@plexus/features` — see [packages/README.md](../../packages/README.md).

## Dev login

After backend `seed-dev`: `admin` / `admin`.

## Future: standalone repo

Move with `packages/{api,ui,features}` (or consume published packages). Keep CSP `connect-src` and CORS aligned with the API host.
