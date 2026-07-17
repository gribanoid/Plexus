.DEFAULT_GOAL := help

# ─── Colors ───────────────────────────────────────────────────────────────────
BOLD  := \033[1m
GREEN := \033[32m
CYAN  := \033[36m
RESET := \033[0m

# Project paths (future split: these become separate repos)
BACKEND  := backend
WEB      := web
DESKTOP  := apps/plexus-desktop
IOS      := apps/plexus-ios
ANDROID  := apps/plexus-android

COMPOSE := docker compose -p plexus-infra -f infra/docker/docker-compose.yml -f infra/docker/docker-compose.override.yml

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
	@node_ver=$$(node -e "process.stdout.write(process.version.replace('v',''))" 2>/dev/null) && \
	  node_major=$$(echo $$node_ver | cut -d. -f1) && \
	  [ "$$node_major" -ge 20 ] && \
	  echo "  $(GREEN)✓$(RESET) Node.js $$node_ver" || \
	  (echo "  ✗ Node.js ≥ 20 required — https://nodejs.org" && exit 1)
	@npm_ver=$$(npm --version 2>/dev/null) && \
	  npm_major=$$(echo $$npm_ver | cut -d. -f1) && \
	  [ "$$npm_major" -ge 10 ] && \
	  echo "  $(GREEN)✓$(RESET) npm $$npm_ver" || \
	  (echo "  ✗ npm ≥ 10 required — run: npm install -g npm@latest" && exit 1)
	@go version > /dev/null 2>&1 && \
	  echo "  $(GREEN)✓$(RESET) $$(go version | awk '{print $$3}')" || \
	  (echo "  ✗ Go ≥ 1.22 required — https://go.dev/dl" && exit 1)
	@docker --version > /dev/null 2>&1 && \
	  echo "  $(GREEN)✓$(RESET) Docker $$(docker --version | awk '{print $$3}' | tr -d ',')" || \
	  (echo "  ✗ Docker required — https://docs.docker.com/get-docker" && exit 1)
	@if [ "$$(uname)" = "Darwin" ]; then \
	  xcodebuild -version > /dev/null 2>&1 && \
	    echo "  $(GREEN)✓$(RESET) Xcode $$(xcodebuild -version 2>/dev/null | head -1 | awk '{print $$2}')" || \
	    echo "  $(CYAN)⚠$(RESET)  Xcode not found — iOS builds unavailable"; \
	fi
	@java_ver=$$(java -version 2>&1 | head -1 | sed 's/.*version "\([0-9]*\).*/\1/') 2>/dev/null; \
	  [ "$$java_ver" -ge 17 ] 2>/dev/null && \
	    echo "  $(GREEN)✓$(RESET) Java $$java_ver" || \
	    echo "  $(CYAN)⚠$(RESET)  Java 17+ not found — Android builds unavailable (brew install --cask temurin@17)"

deps-js: ## Install JS/TS dependencies (web + desktop + packages)
	@echo "$(BOLD)Installing JS dependencies…$(RESET)"
	ELECTRON_SKIP_BINARY_DOWNLOAD=1 npm install
	@echo "  $(GREEN)✓$(RESET) JS dependencies installed"

deps-go: ## Download Go modules (backend)
	@$(MAKE) -C $(BACKEND) deps

deps-ios: ## Resolve iOS Swift packages
	@$(MAKE) -C $(IOS) deps

deps-android: ## Prepare Android toolchain + Gradle deps
	@$(MAKE) -C $(ANDROID) deps

# ─── infra / orchestration ────────────────────────────────────────────────────
.PHONY: infra infra-down

infra: ## Start infrastructure (PostgreSQL, Redis, Meilisearch, MinIO)
	$(COMPOSE) up -d postgres redis meilisearch minio
	@echo "$(GREEN)✓ Infrastructure started$(RESET)"
	@echo "  Postgres    → localhost:5432"
	@echo "  Redis       → localhost:6379"
	@echo "  Meilisearch → http://localhost:7700"
	@echo "  MinIO       → http://localhost:9000  (console: http://localhost:9001)"

infra-down: ## Stop infrastructure containers
	$(COMPOSE) down

# ─── backend (delegates) ──────────────────────────────────────────────────────
.PHONY: dev-backend migrate migrate-down seed-dev build-backend

dev-backend: ## Run Go backend → :8080
	@$(MAKE) -C $(BACKEND) dev

migrate: ## Apply DB migrations
	@$(MAKE) -C $(BACKEND) migrate

migrate-down: ## Roll back last migration
	@$(MAKE) -C $(BACKEND) migrate-down

seed-dev: ## Seed admin + demo project (development only)
	@$(MAKE) -C $(BACKEND) seed-dev

build-backend: ## Build plexus-server + plexus-worker → dist/
	@$(MAKE) -C $(BACKEND) build

# ─── web (delegates) ──────────────────────────────────────────────────────────
.PHONY: dev-web dev-web-stop start-web build-web open-web

dev-web: ## Run Next.js → http://localhost:3000
	@$(MAKE) -C $(WEB) dev

dev-web-stop: ## Stop process on port 3000
	@$(MAKE) -C $(WEB) stop

start-web: ## Production Next.js start
	@$(MAKE) -C $(WEB) start

build-web: ## Build Next.js for production
	@$(MAKE) -C $(WEB) build

open-web: ## Open web app in browser
	open http://localhost:3000

# ─── desktop (delegates) ──────────────────────────────────────────────────────
.PHONY: dev-desktop preview-desktop build-desktop

dev-desktop: ## Run Electron desktop app
	@$(MAKE) -C $(DESKTOP) dev

preview-desktop: ## Preview built Electron app
	@$(MAKE) -C $(DESKTOP) preview

build-desktop: ## Build Electron distributable
	@$(MAKE) -C $(DESKTOP) package

# ─── mobile (delegates) ───────────────────────────────────────────────────────
.PHONY: build-ios build-android run-android

build-ios: ## Build iOS for Simulator
	@$(MAKE) -C $(IOS) build

build-android: ## Assemble Android debug APK
	@$(MAKE) -C $(ANDROID) build

run-android: ## Build + install Android debug on device/emulator
	@$(MAKE) -C $(ANDROID) run

# ─── combined ─────────────────────────────────────────────────────────────────
.PHONY: build dev

build: build-backend build-web ## Build backend + web for production

dev: infra ## Start infra + backend + web in parallel (tmux) or print instructions
	@command -v tmux > /dev/null 2>&1 && { \
	  tmux new-session -d -s plexus-dev -n backend 'make dev-backend'; \
	  tmux new-window -t plexus-dev -n web 'make dev-web'; \
	  tmux new-window -t plexus-dev -n desktop 'make dev-desktop'; \
	  tmux select-window -t plexus-dev:web; \
	  tmux attach -t plexus-dev; \
	} || { \
	  echo "$(CYAN)tmux not found — open terminals manually:$(RESET)"; \
	  echo "  1) make dev-backend"; \
	  echo "  2) make dev-web"; \
	  echo "  3) make dev-desktop"; \
	}

# ─── quality ──────────────────────────────────────────────────────────────────
.PHONY: lint typecheck fmt

lint: ## Lint JS workspaces + go vet
	npm run lint --workspaces --if-present
	cd $(BACKEND) && go vet ./...

typecheck: ## Typecheck TypeScript workspaces
	npm run typecheck --workspaces --if-present

fmt: ## Format JS + Go
	npm run format --if-present
	@$(MAKE) -C $(BACKEND) fmt

# ─── clean ────────────────────────────────────────────────────────────────────
.PHONY: clean

clean: ## Remove build artifacts and node_modules
	rm -rf node_modules apps/*/node_modules packages/*/node_modules
	rm -rf $(WEB)/.next $(DESKTOP)/out dist
	@echo "$(GREEN)✓ Clean$(RESET)"
