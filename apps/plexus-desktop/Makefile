.DEFAULT_GOAL := help

BOLD  := \033[1m
GREEN := \033[32m
CYAN  := \033[36m
RESET := \033[0m

ROOT := $(abspath ../..)
WORKSPACE := --workspace=@plexus/desktop

.PHONY: help deps dev build preview package typecheck lint

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(BOLD)Usage:$(RESET)\n  make $(CYAN)<target>$(RESET)\n\n$(BOLD)Targets:$(RESET)\n"} \
	  /^[a-zA-Z_-]+:.*?##/ { printf "  $(CYAN)%-14s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

deps: ## Install JS deps (and Electron binary if needed)
	cd $(ROOT) && ELECTRON_SKIP_BINARY_DOWNLOAD=1 npm install
	@echo "$(GREEN)✓$(RESET) Dependencies installed"
	@echo "  If Electron fails to start: cd $(ROOT) && node node_modules/electron/install.js"

dev: ## Electron + Vite renderer (unset ELECTRON_RUN_AS_NODE)
	cd $(ROOT) && env -u ELECTRON_RUN_AS_NODE ELECTRON_SKIP_BINARY_DOWNLOAD=1 npm run dev $(WORKSPACE)

build: ## electron-vite production build → out/
	cd $(ROOT) && npm run build $(WORKSPACE)

preview: ## Preview production build in Electron
	cd $(ROOT) && env -u ELECTRON_RUN_AS_NODE npm run preview $(WORKSPACE)

package: ## Build installers (electron-builder)
	cd $(ROOT) && ELECTRON_SKIP_BINARY_DOWNLOAD=1 npm run package $(WORKSPACE)

typecheck: ## TypeScript check
	cd $(ROOT) && npm run typecheck $(WORKSPACE)

lint: ## ESLint
	cd $(ROOT) && npm run lint $(WORKSPACE)
