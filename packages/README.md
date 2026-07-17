# Shared packages

JavaScript/TypeScript libraries shared by **web** and **desktop** only. Mobile apps do not use these packages.

| Package | Name | Role |
|---|---|---|
| [`api/`](api/) | `@plexus/api` | `configureApi`, auth helpers, routes, fetch, WebSocket, React Query hooks |
| [`ui/`](ui/) | `@plexus/ui` | Shared React UI primitives (buttons, badges, avatars, …) |
| [`features/`](features/) | `@plexus/features` | Feature modules (create/edit issue, project dialog, search, notifications) |
| `api-client/` | — | Placeholder for generated OpenAPI client (unused) |

## Consumers

- [`web`](../web/)
- [`apps/plexus-desktop`](../apps/plexus-desktop/)

Dependency graph: `features` → `api` + `ui`; apps → all three.

## Development

Installed via the monorepo root (`npm workspaces`: `packages/*`). Source is consumed directly (TypeScript entrypoints); no separate publish step today.

```bash
# from repo root
npm run typecheck --workspaces --if-present
```

## Future: standalone / split

When web and desktop leave the monorepo, either:

1. **Ship packages with web/desktop** (same repo or git submodule), or
2. **Publish `@plexus/*`** to a private registry and version against `backend/openapi.yaml`.

iOS and Android remain independent and only need the OpenAPI contract from the backend.
