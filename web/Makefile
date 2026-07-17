.DEFAULT_GOAL := help

BOLD  := \033[1m
GREEN := \033[32m
CYAN  := \033[36m
RESET := \033[0m

# Monorepo root. When this app is a standalone repo, set ROOT := $(CURDIR) and drop WORKSPACE.
ROOT := $(abspath ..)
WORKSPACE := --workspace=@plexus/web

.PHONY: help deps dev stop build start typecheck lint test-e2e

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(BOLD)Usage:$(RESET)\n  make $(CYAN)<target>$(RESET)\n\n$(BOLD)Targets:$(RESET)\n"} \
	  /^[a-zA-Z_-]+:.*?##/ { printf "  $(CYAN)%-14s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

deps: ## Install JS deps from monorepo root
	cd $(ROOT) && ELECTRON_SKIP_BINARY_DOWNLOAD=1 npm install
	@echo "$(GREEN)✓$(RESET) Dependencies installed"

dev: ## Next.js dev server → http://localhost:3000
	@if lsof -ti:3000 > /dev/null 2>&1; then \
	  echo "$(CYAN)⚠ Port 3000 in use. Run: make stop$(RESET)"; exit 1; \
	fi
	cd $(ROOT) && npm run dev $(WORKSPACE)

stop: ## Stop process on port 3000
	@if lsof -ti:3000 > /dev/null 2>&1; then \
	  lsof -ti:3000 | xargs kill -9; \
	  echo "$(GREEN)✓$(RESET) Stopped process on port 3000"; \
	else \
	  echo "$(CYAN)No process on port 3000$(RESET)"; \
	fi

build: ## Production Next.js build
	cd $(ROOT) && npm run build $(WORKSPACE)

start: ## Start production server (after build)
	cd $(ROOT) && npm run start $(WORKSPACE)

typecheck: ## TypeScript check
	cd $(ROOT) && npm run typecheck $(WORKSPACE)

lint: ## ESLint
	cd $(ROOT) && npm run lint $(WORKSPACE)

test-e2e: ## Playwright e2e
	cd $(ROOT) && npm run test:e2e $(WORKSPACE)
