.DEFAULT_GOAL := help

BOLD  := \033[1m
GREEN := \033[32m
CYAN  := \033[36m
RESET := \033[0m

PROJECT := Plexus/Plexus.xcodeproj
SCHEME  := Plexus

.PHONY: help deps build open

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(BOLD)Usage:$(RESET)\n  make $(CYAN)<target>$(RESET)\n\n$(BOLD)Targets:$(RESET)\n"} \
	  /^[a-zA-Z_-]+:.*?##/ { printf "  $(CYAN)%-10s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

deps: ## Resolve Swift Package Manager dependencies
	xcodebuild -resolvePackageDependencies \
	  -project $(PROJECT) \
	  -scheme $(SCHEME)
	@echo "$(GREEN)✓$(RESET) iOS dependencies resolved"

build: ## Build for iOS Simulator (no code signing)
	@echo "$(BOLD)Building iOS app…$(RESET)"
	xcodebuild build \
	  -project $(PROJECT) \
	  -scheme $(SCHEME) \
	  -destination 'generic/platform=iOS Simulator' \
	  -configuration Debug \
	  CODE_SIGNING_ALLOWED=NO
	@echo "$(GREEN)✓$(RESET) iOS build complete"

open: ## Open Xcode project
	open $(PROJECT)
