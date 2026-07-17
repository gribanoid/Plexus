.DEFAULT_GOAL := help

BOLD  := \033[1m
GREEN := \033[32m
CYAN  := \033[36m
RESET := \033[0m

# Monorepo root (parent). When this directory is a standalone repo, set ROOT := $(CURDIR)
ROOT := $(abspath ..)
COMPOSE := docker compose -p plexus-infra -f $(ROOT)/infra/docker/docker-compose.yml -f $(ROOT)/infra/docker/docker-compose.override.yml

.PHONY: help deps infra infra-down dev build migrate migrate-down seed-dev test fmt

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(BOLD)Usage:$(RESET)\n  make $(CYAN)<target>$(RESET)\n\n$(BOLD)Targets:$(RESET)\n"} \
	  /^[a-zA-Z_-]+:.*?##/ { printf "  $(CYAN)%-18s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

deps: ## Download Go modules
	go mod download
	@echo "$(GREEN)✓$(RESET) Go modules downloaded"

infra: ## Start infra (Postgres, Redis, Meilisearch, MinIO) via monorepo compose
	$(COMPOSE) up -d postgres redis meilisearch minio
	@echo "$(GREEN)✓ Infrastructure started$(RESET)"

infra-down: ## Stop infrastructure containers
	$(COMPOSE) down

dev: ## Run API server (applies migrations on startup)
	@[ -f .env ] || (cp .env.example .env && \
	  echo "$(CYAN)Created backend/.env from example$(RESET)")
	go run ./cmd/server

build: ## Build plexus-server and plexus-worker into ../dist (or ./dist if standalone)
	@mkdir -p $(ROOT)/dist
	go build -ldflags="-s -w" -o $(ROOT)/dist/plexus-server ./cmd/server
	go build -ldflags="-s -w" -o $(ROOT)/dist/plexus-worker ./cmd/worker
	@echo "$(GREEN)✓$(RESET) $(ROOT)/dist/plexus-server + plexus-worker"

migrate: ## Apply pending goose migrations
	@set -a && [ -f .env ] && . ./.env; set +a; \
	  go run github.com/pressly/goose/v3/cmd/goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down: ## Roll back the last migration
	@set -a && [ -f .env ] && . ./.env; set +a; \
	  go run github.com/pressly/goose/v3/cmd/goose -dir migrations postgres "$$DATABASE_URL" down

seed-dev: ## Seed admin + demo project (development only; blocked when ENV=production)
	@set -a && [ -f .env ] && . ./.env; set +a; \
	  if [ "$${ENV:-development}" = "production" ]; then \
	    echo "✗ seed-dev is forbidden when ENV=production"; exit 1; \
	  fi; \
	  psql "$$DATABASE_URL" -f scripts/seed-dev.sql
	@echo "$(GREEN)✓ Dev seed applied$(RESET)"
	@echo "  Login: admin@plexus.local / admin  (or admin / admin)"

test: ## Run Go tests
	go test ./...

fmt: ## Format Go sources
	gofmt -w .
