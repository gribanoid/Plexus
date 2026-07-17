# Plexus Android

Jetpack Compose client for Plexus. Uses a hand-written Retrofit API (not `@plexus/api`).

**Stack:** Kotlin · Jetpack Compose · Hilt · Retrofit · Gradle

## Prerequisites

- Java 17+ (Temurin recommended)
- Android SDK (Android Studio)
- Running [backend](../../backend/README.md) on the host

## Setup

```bash
# monorepo root
make infra && make migrate && make seed-dev
make dev-backend

make -C apps/plexus-android deps
# Start an AVD in Android Studio, then:
make -C apps/plexus-android run
```

Or from root: `make deps-android` / `make run-android`.

### SDK path

`make deps` writes `local.properties` from `ANDROID_HOME` or default paths:

- macOS: `~/Library/Android/sdk`
- Linux: `~/Android/Sdk`

Or copy [`local.properties.example`](local.properties.example).

## Commands

| Target | Description |
|---|---|
| `make deps` | Java check, `local.properties`, Gradle sync |
| `make build` | `assembleDebug` APK |
| `make install` / `run` | Build + install on emulator/device |

## API URL

Debug builds use `http://10.0.2.2:8080/api/v1` (emulator → host loopback). See `app/build.gradle.kts` (`BuildConfig.API_BASE_URL`).

**USB device:**

```bash
adb reverse tcp:8080 tcp:8080
# then use http://127.0.0.1:8080/api/v1 in debug BuildConfig, or your LAN IP
```

## Dev login

After backend `seed-dev`: `admin` / `admin`.

## Contract

Align with [`backend/openapi.yaml`](../../backend/openapi.yaml). Client: `app/.../data/api/PlexusApi.kt`.

## Future: standalone repo

No dependency on `packages/*`. Move this directory as its own repo; keep Gradle wrapper and this Makefile. Point `API_BASE_URL` at any Plexus API.
