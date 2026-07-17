.DEFAULT_GOAL := help

BOLD  := \033[1m
GREEN := \033[32m
CYAN  := \033[36m
RESET := \033[0m

# Prefer Temurin 17 when JAVA_HOME is unset
export JAVA_HOME ?= $(shell \
  if [ -d "/Library/Java/JavaVirtualMachines/temurin-17.jdk/Contents/Home" ]; then \
    echo "/Library/Java/JavaVirtualMachines/temurin-17.jdk/Contents/Home"; \
  elif command -v /usr/libexec/java_home >/dev/null 2>&1; then \
    /usr/libexec/java_home -v 17 2>/dev/null || /usr/libexec/java_home 2>/dev/null; \
  fi)

.PHONY: help deps build run install _ensure-local-properties

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(BOLD)Usage:$(RESET)\n  make $(CYAN)<target>$(RESET)\n\n$(BOLD)Targets:$(RESET)\n"} \
	  /^[a-zA-Z_-]+:.*?##/ { printf "  $(CYAN)%-10s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

deps: _ensure-local-properties ## Ensure Java 17+, write local.properties, sync Gradle
	@java_ok=0; \
	  if command -v java >/dev/null 2>&1; then \
	    java_ver=$$(java -version 2>&1 | head -1 | sed 's/.*version "\([0-9]*\).*/\1/'); \
	    if [ "$$java_ver" -ge 17 ] 2>/dev/null; then java_ok=1; fi; \
	  fi; \
	  if [ "$$java_ok" != "1" ]; then \
	    if [ "$$(uname)" = "Darwin" ] && command -v brew >/dev/null 2>&1; then \
	      echo "Installing Temurin JDK 17 via Homebrew…"; \
	      brew install --cask temurin@17; \
	    else \
	      echo "✗ Java 17+ required (brew install --cask temurin@17)"; exit 1; \
	    fi; \
	  fi
	@echo "$(GREEN)✓$(RESET) $$(java -version 2>&1 | head -1)"
	./gradlew :app:dependencies --configuration debugRuntimeClasspath -q
	@echo "$(GREEN)✓$(RESET) Android dependencies synced"

_ensure-local-properties:
	@if [ ! -f local.properties ]; then \
	  if [ -n "$$ANDROID_HOME" ]; then \
	    echo "sdk.dir=$$ANDROID_HOME" > local.properties; \
	  elif [ -d "$$HOME/Library/Android/sdk" ]; then \
	    echo "sdk.dir=$$HOME/Library/Android/sdk" > local.properties; \
	  elif [ -d "$$HOME/Android/Sdk" ]; then \
	    echo "sdk.dir=$$HOME/Android/Sdk" > local.properties; \
	  else \
	    echo "$(CYAN)⚠$(RESET) No SDK — copy local.properties.example → local.properties"; \
	  fi; \
	fi

build: _ensure-local-properties ## Assemble debug APK
	./gradlew assembleDebug
	@echo "$(GREEN)✓$(RESET) app/build/outputs/apk/debug/app-debug.apk"

install: _ensure-local-properties ## installDebug on connected device/emulator
	./gradlew installDebug
	@echo "$(GREEN)✓$(RESET) Installed debug build"

run: install ## Alias for install (build + install)
	@echo "  Emulator API: http://10.0.2.2:8080 — start backend on the host"
	@echo "  Login after seed: admin / admin"
