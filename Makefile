.DEFAULT_GOAL := help

# ─── Colors ───────────────────────────────────────────────────────────────────
BOLD  := \033[1m
GREEN := \033[32m
CYAN  := \033[36m
RESET := \033[0m

# ─── Helpers ──────────────────────────────────────────────────────────────────
.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(BOLD)Usage:$(RESET)\n  make $(CYAN)<target>$(RESET)\n\n$(BOLD)Targets:$(RESET)\n"} \
	  /^[a-zA-Z_-]+:.*?##/ { printf "  $(CYAN)%-22s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# ─── deps ─────────────────────────────────────────────────────────────────────
.PHONY: deps deps-check deps-js deps-go deps-ios deps-android

deps: deps-check deps-js deps-go ## Install all dependencies (JS + Go)
	@echo "$(GREEN)✓ All dependencies installed$(RESET)"

deps-check: ## Verify required tools are installed
	@echo "$(BOLD)Checking required tools…$(RESET)"

	@# Node.js ≥ 20
	@node_ver=$$(node -e "process.stdout.write(process.version.replace('v',''))" 2>/dev/null) && \
	  node_major=$$(echo $$node_ver | cut -d. -f1) && \
	  [ "$$node_major" -ge 20 ] && \
	  echo "  $(GREEN)✓$(RESET) Node.js $$node_ver" || \
	  (echo "  ✗ Node.js ≥ 20 required — https://nodejs.org" && exit 1)

	@# npm ≥ 10
	@npm_ver=$$(npm --version 2>/dev/null) && \
	  npm_major=$$(echo $$npm_ver | cut -d. -f1) && \
	  [ "$$npm_major" -ge 10 ] && \
	  echo "  $(GREEN)✓$(RESET) npm $$npm_ver" || \
	  (echo "  ✗ npm ≥ 10 required — run: npm install -g npm@latest" && exit 1)

	@# Go ≥ 1.22
	@go version > /dev/null 2>&1 && \
	  echo "  $(GREEN)✓$(RESET) $$(go version | awk '{print $$3}')" || \
	  (echo "  ✗ Go ≥ 1.22 required — https://go.dev/dl" && exit 1)

	@# Docker CLI (daemon может быть не запущен — нормально на этапе deps-check)
	@docker --version > /dev/null 2>&1 && \
	  echo "  $(GREEN)✓$(RESET) Docker $$(docker --version | awk '{print $$3}' | tr -d ',')" || \
	  (echo "  ✗ Docker required — https://docs.docker.com/get-docker" && exit 1)

	@# Xcode (macOS only, optional — warns but doesn't fail)
	@if [ "$$(uname)" = "Darwin" ]; then \
	  xcodebuild -version > /dev/null 2>&1 && \
	    echo "  $(GREEN)✓$(RESET) Xcode $$(xcodebuild -version 2>/dev/null | head -1 | awk '{print $$2}')" || \
	    echo "  $(CYAN)⚠$(RESET)  Xcode not found — iOS builds unavailable (App Store: https://apps.apple.com/app/xcode/id497799835)"; \
	fi

	@# Java 17+ (optional — warns but doesn't fail)
	@java_ver=$$(java -version 2>&1 | head -1 | sed 's/.*version "\([0-9]*\).*/\1/') 2>/dev/null; \
	  [ "$$java_ver" -ge 17 ] 2>/dev/null && \
	    echo "  $(GREEN)✓$(RESET) Java $$java_ver" || \
	    echo "  $(CYAN)⚠$(RESET)  Java 17+ not found — Android builds unavailable (brew install --cask temurin@17)"

deps-js: ## Install JS/TS dependencies (web + desktop)
	@echo "$(BOLD)Installing JS dependencies…$(RESET)"
	ELECTRON_SKIP_BINARY_DOWNLOAD=1 npm install
	@echo "  $(GREEN)✓$(RESET) JS dependencies installed"

deps-go: ## Download Go modules (backend)
	@echo "$(BOLD)Downloading Go modules…$(RESET)"
	cd backend && go mod download
	@echo "  $(GREEN)✓$(RESET) Go modules downloaded"

deps-ios: ## Install iOS Swift Package Manager dependencies (requires Xcode)
	@echo "$(BOLD)Resolving iOS Swift packages…$(RESET)"
	xcodebuild -resolvePackageDependencies \
	  -project apps/ios/Plexus/Plexus.xcodeproj \
	  -scheme Plexus
	@echo "  $(GREEN)✓$(RESET) iOS dependencies resolved"

deps-android: ## Sync Android Gradle dependencies (requires Java 17+)
	@echo "$(BOLD)Syncing Android Gradle dependencies…$(RESET)"
	cd apps/android && ./gradlew :app:dependencies --configuration releaseRuntimeClasspath -q
	@echo "  $(GREEN)✓$(RESET) Android dependencies synced"

# ─── dev ──────────────────────────────────────────────────────────────────────
COMPOSE := docker compose -p plexus-infra -f infra/docker/docker-compose.yml -f infra/docker/docker-compose.override.yml

.PHONY: infra infra-down dev-backend dev-web dev-web-stop dev-desktop dev

infra: ## Start infrastructure (PostgreSQL, Redis, Meilisearch, MinIO)
	$(COMPOSE) up -d postgres redis meilisearch minio
	@echo "$(GREEN)✓ Infrastructure started$(RESET)"
	@echo "  Postgres    → localhost:5432"
	@echo "  Redis       → localhost:6379"
	@echo "  Meilisearch → http://localhost:7700"
	@echo "  MinIO       → http://localhost:9000  (console: http://localhost:9001)"

infra-down: ## Stop infrastructure containers
	$(COMPOSE) down

dev-backend: ## Run Go backend (applies migrations on startup)
	@[ -f backend/.env ] || (cp backend/.env.example backend/.env && \
	  echo "$(CYAN)Created backend/.env from example — edit it before running in production$(RESET)")
	cd backend && go run cmd/server/main.go

dev-web: _check-node-modules ## Run Next.js web app → http://localhost:3000
	@if lsof -ti:3000 > /dev/null 2>&1; then \
	  echo "$(CYAN)⚠ Port 3000 is already in use. Run \`make dev-web-stop\` first, or stop the other process.$(RESET)"; \
	  exit 1; \
	fi
	npm run dev --workspace=apps/web

dev-web-stop: ## Stop the process listening on port 3000 (Next.js dev server)
	@if lsof -ti:3000 > /dev/null 2>&1; then \
	  lsof -ti:3000 | xargs kill -9; \
	  echo "$(GREEN)✓ Stopped process on port 3000$(RESET)"; \
	else \
	  echo "$(CYAN)No process found on port 3000$(RESET)"; \
	fi

dev-desktop: _check-node-modules ## Run Electron desktop app
	ELECTRON_SKIP_BINARY_DOWNLOAD=1 npm run dev --workspace=apps/desktop

start-web: _check-node-modules ## Run Next.js in production mode (requires build-web first)
	npm run start --workspace=apps/web

preview-desktop: _check-node-modules ## Preview built Electron app
	npm run preview --workspace=apps/desktop

open-web: ## Open web app in browser (assumes dev-web is running)
	open http://localhost:3000

dev: infra ## Start infra + backend + web in parallel (requires tmux or runs sequentially)
	@command -v tmux > /dev/null 2>&1 && { \
	  tmux new-session -d -s plexus-dev -n backend 'make dev-backend'; \
	  tmux new-window -t plexus-dev -n web 'make dev-web'; \
	  tmux new-window -t plexus-dev -n desktop 'make dev-desktop'; \
	  tmux select-window -t plexus-dev:web; \
	  tmux attach -t plexus-dev; \
	} || { \
	  echo "$(CYAN)tmux not found — открой 3 терминала и запусти каждый вручную:$(RESET)"; \
	  echo "  1) make dev-backend"; \
	  echo "  2) make dev-web"; \
	  echo "  3) make dev-desktop"; \
	  echo ""; \
	  echo "$(CYAN)Или установи tmux: brew install tmux$(RESET)"; \
	}

# Вспомогательная цель — проверяет наличие node_modules
.PHONY: _check-node-modules
_check-node-modules:
	@[ -d node_modules ] || { \
	  echo "$(CYAN)node_modules не найден — запускаю npm install…$(RESET)"; \
	  ELECTRON_SKIP_BINARY_DOWNLOAD=1 npm install; \
	}

# ─── build ────────────────────────────────────────────────────────────────────
.PHONY: build build-backend build-web build-desktop build-ios build-android

build: build-backend build-web ## Build backend + web for production

build-backend: ## Build Go binary → dist/plexus-server
	cd backend && go build -ldflags="-s -w" -o ../dist/plexus-server ./cmd/server

build-web: ## Build Next.js for production
	npm run build --workspace=apps/web

build-desktop: ## Build Electron distributable
	ELECTRON_SKIP_BINARY_DOWNLOAD=1 npm run package --workspace=apps/desktop

build-ios: ## Build iOS app for simulator (requires Xcode)
	@echo "$(BOLD)Building iOS app…$(RESET)"
	xcodebuild build \
	  -project apps/ios/Plexus/Plexus.xcodeproj \
	  -scheme Plexus \
	  -destination 'generic/platform=iOS Simulator' \
	  -configuration Debug \
	  CODE_SIGNING_ALLOWED=NO
	@echo "  $(GREEN)✓$(RESET) iOS build complete"

build-android: ## Build Android debug APK (requires Java 17+ and Android SDK)
	@echo "$(BOLD)Building Android app…$(RESET)"
	cd apps/android && ./gradlew assembleDebug
	@echo "  $(GREEN)✓$(RESET) Android APK → apps/android/app/build/outputs/apk/debug/app-debug.apk"

# ─── db ───────────────────────────────────────────────────────────────────────
.PHONY: migrate migrate-down seed-dev

migrate: ## Apply all pending DB migrations
	@set -a && [ -f backend/.env ] && . backend/.env; set +a; \
	  cd backend && go run github.com/pressly/goose/v3/cmd/goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down: ## Roll back the last DB migration
	@set -a && [ -f backend/.env ] && . backend/.env; set +a; \
	  cd backend && go run github.com/pressly/goose/v3/cmd/goose -dir migrations postgres "$$DATABASE_URL" down

seed-dev: ## Seed development data (admin user, demo project) — development only
	@set -a && [ -f backend/.env ] && . backend/.env; set +a; \
	  psql "$$DATABASE_URL" -f backend/scripts/seed-dev.sql
	@echo "$(GREEN)✓ Dev seed applied$(RESET)"
	@echo "  Login: admin@plexus.local / admin"

# ─── lint & typecheck ─────────────────────────────────────────────────────────
.PHONY: lint typecheck fmt

lint: ## Lint all workspaces
	npm run lint --workspaces --if-present
	cd backend && go vet ./...

typecheck: ## Typecheck TypeScript
	npm run typecheck --workspaces --if-present

fmt: ## Format all code
	npm run format --if-present
	cd backend && gofmt -w .

# ─── clean ────────────────────────────────────────────────────────────────────
.PHONY: clean

clean: ## Remove build artifacts and node_modules
	rm -rf node_modules apps/*/node_modules packages/*/node_modules
	rm -rf apps/web/.next apps/desktop/out dist
	@echo "$(GREEN)✓ Clean$(RESET)"
