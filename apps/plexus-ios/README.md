# Plexus iOS

SwiftUI client for Plexus. Uses a hand-written networking layer (not `@plexus/api`).

**Stack:** SwiftUI · Swift Concurrency · Keychain · Xcode 15+

## Prerequisites

- macOS with Xcode 15+
- Running [backend](../../backend/README.md) on the Mac (`:8080`)

## Setup (Simulator)

```bash
# monorepo root
make infra && make migrate && make seed-dev
make dev-backend

make -C apps/plexus-ios deps
make -C apps/plexus-ios open
# Xcode → pick iPhone Simulator → Run ▶
```

Or: `make deps-ios` / `make build-ios` from the repo root.

## Commands

| Target | Description |
|---|---|
| `make deps` | Resolve SPM packages |
| `make build` | Debug build for iOS Simulator |
| `make open` | Open `Plexus.xcodeproj` in Xcode |

## API URL

[`Plexus/Info.plist`](Plexus/Info.plist) key `PLEXUS_API_URL` (default `http://127.0.0.1:8080/api/v1`).

Simulator reaches the host Mac via `127.0.0.1`. Local HTTP is allowed via `NSAllowsLocalNetworking`.

**Physical device:** set `PLEXUS_API_URL` to your Mac’s LAN IP (e.g. `http://192.168.x.x:8080/api/v1`) and keep the device on the same network.

## Dev login

After backend `seed-dev`: `admin` / `admin` (or `admin@plexus.local` / `admin`).

## Contract

Align with [`backend/openapi.yaml`](../../backend/openapi.yaml). Client code: `Plexus/Shared/Network/APIClient.swift`.

## Future: standalone repo

This app has no dependency on `packages/*`. Clone as its own repo with this Makefile/README; point `PLEXUS_API_URL` at any Plexus API and keep OpenAPI in sync.
