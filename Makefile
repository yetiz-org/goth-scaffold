SHELL := /bin/bash

PROJECT_NAME := scaffold
EVALUATE_DIR := evaluate
COMPOSE_FILE := $(EVALUATE_DIR)/docker-compose.yml
COMPOSE      ?= docker compose
CONFIG_FILE  ?= $(EVALUATE_DIR)/config.yaml.local

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    BINARY_NAME := $(PROJECT_NAME)-darwin
else
    BINARY_NAME := $(PROJECT_NAME)-amd64
endif

CGO_ENABLED      ?= 0
GO_BUILD_FLAGS   := -trimpath
GO_BUILD_LDFLAGS := -s -w

# ─── Colours ─────────────────────────────────────────────────────────────────
RED    := $(shell printf '\033[0;31m')
GREEN  := $(shell printf '\033[0;32m')
YELLOW := $(shell printf '\033[0;33m')
BLUE   := $(shell printf '\033[0;34m')
NC     := $(shell printf '\033[0m')

# ─── Default target ──────────────────────────────────────────────────────────
.PHONY: help
help: ## Show this help
	@echo "$(BLUE)goth-scaffold Local Development$(NC)"
	@echo "================================"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "  $(YELLOW)%-30s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ─── Local Environment (Docker Compose) ──────────────────────────────────────
.PHONY: local-env-setup
local-env-setup: ## Create config.yaml.local from example (first-time setup)
	@if [ ! -f $(CONFIG_FILE) ]; then \
		cp $(EVALUATE_DIR)/config.yaml.local.example $(CONFIG_FILE); \
		echo "$(GREEN)[OK]$(NC) Created $(CONFIG_FILE) — edit as needed"; \
	else \
		echo "$(YELLOW)[SKIP]$(NC) $(CONFIG_FILE) already exists"; \
	fi

.PHONY: local-env-start
local-env-start: ## Start all Docker services
	@echo "$(BLUE)[INFO]$(NC) Starting services..."
	$(COMPOSE) -f $(COMPOSE_FILE) up -d
	@echo "$(GREEN)[OK]$(NC) Services started"

.PHONY: local-env-status
local-env-status: ## Show Docker service status
	$(COMPOSE) -f $(COMPOSE_FILE) ps

.PHONY: local-env-logs
local-env-logs: ## Follow Docker service logs (Ctrl-C to exit)
	$(COMPOSE) -f $(COMPOSE_FILE) logs -f

.PHONY: local-env-stop
local-env-stop: ## Stop all Docker services
	@echo "$(BLUE)[INFO]$(NC) Stopping services..."
	$(COMPOSE) -f $(COMPOSE_FILE) down
	@echo "$(GREEN)[OK]$(NC) Services stopped"

.PHONY: local-env-clean
local-env-clean: ## Destroy all local data volumes and config (destructive)
	@echo "$(YELLOW)[WARN]$(NC) Destroying all local data (MySQL / Redis / Cassandra)..."
	$(COMPOSE) -f $(COMPOSE_FILE) down -v
	@if [ -d "$(EVALUATE_DIR)/_run" ]; then \
		docker run --user root --rm -v "$$(pwd)/$(EVALUATE_DIR)/_run:/data" redis:7-alpine \
			sh -c "rm -rf /data/*" 2>/dev/null || true; \
		rm -rf $(EVALUATE_DIR)/_run; \
	fi
	@rm -f $(EVALUATE_DIR)/config.yaml.local
	@echo "$(GREEN)[OK]$(NC) Environment cleaned"

# ─── Service Connections ──────────────────────────────────────────────────────
.PHONY: local-database-connect
local-database-connect: ## Open MySQL CLI
	docker exec -it $(PROJECT_NAME)-mysql mysql -u root -proot

.PHONY: local-redis-connect
local-redis-connect: ## Open Redis CLI
	docker exec -it $(PROJECT_NAME)-redis redis-cli

.PHONY: local-cassandra-connect
local-cassandra-connect: ## Open Cassandra CQL shell
	docker exec -it $(PROJECT_NAME)-cassandra cqlsh

# ─── Build ───────────────────────────────────────────────────────────────────
.PHONY: build
build: ## Compile binary for current OS
	@echo "$(BLUE)[INFO]$(NC) Building $(BINARY_NAME)..."
	@CGO_ENABLED=$(CGO_ENABLED) go build $(GO_BUILD_FLAGS) -ldflags="$(GO_BUILD_LDFLAGS)" -o $(BINARY_NAME) ./
	@echo "$(GREEN)[OK]$(NC) Build complete: $(BINARY_NAME)"

.PHONY: build-all
build-all: ## Compile binaries for Darwin and Linux/amd64
	@CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build $(GO_BUILD_FLAGS) -ldflags="$(GO_BUILD_LDFLAGS)" -o $(PROJECT_NAME)-darwin ./
	@CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build $(GO_BUILD_FLAGS) -ldflags="$(GO_BUILD_LDFLAGS)" -o $(PROJECT_NAME)-amd64 ./
	@echo "$(GREEN)[OK]$(NC) Build all complete"

.PHONY: clean
clean: ## Remove compiled binaries
	@rm -f $(PROJECT_NAME)-darwin $(PROJECT_NAME)-amd64
	@echo "$(GREEN)[OK]$(NC) Cleaned"

# ─── Run ─────────────────────────────────────────────────────────────────────
.PHONY: local-run
local-run: build ## Build and start API on :8080 (default mode)
	@echo "$(BLUE)[INFO]$(NC) Starting API on :8080 (mode=default)"
	./$(BINARY_NAME) -c $(CONFIG_FILE) -m default

# ─── Service Health Wait ─────────────────────────────────────────────────────
# Internal target: polls until MySQL is accepting connections (max 60 s).
.PHONY: _local-wait-mysql
_local-wait-mysql:
	@echo "$(BLUE)[INFO]$(NC) Waiting for MySQL..."
	@for i in $$(seq 1 30); do \
		docker exec $(PROJECT_NAME)-mysql mysqladmin ping -h localhost --silent 2>/dev/null && \
			echo "$(GREEN)[OK]$(NC) MySQL ready" && exit 0; \
		sleep 2; \
	done; \
	echo "$(RED)[ERR]$(NC) MySQL not ready after 60 s" && exit 1

# ─── Database Operations ─────────────────────────────────────────────────────
# local-db-migration ensures the environment is running before migrating.
# local-db-seed is the full bootstrap: setup → start services → wait → migrate → seed.
# Running `make local-env-clean && make local-db-seed` produces a clean, seeded DB.
# Running `make local-db-reseed` drops and recreates the DB then migrates + seeds.
.PHONY: local-db-migration
local-db-migration: local-env-setup local-env-start _local-wait-mysql build ## Ensure env is up, then run database migrations
	@echo "$(BLUE)[INFO]$(NC) Running migrations..."
	./$(BINARY_NAME) -c $(CONFIG_FILE) -m db_migration
	@echo "$(GREEN)[OK]$(NC) Migrations done"

.PHONY: local-db-seed
local-db-seed: local-db-migration ## Run migrations then seed data
	@echo "$(BLUE)[INFO]$(NC) Running seeds..."
	./$(BINARY_NAME) -c $(CONFIG_FILE) -m db_seed
	@echo "$(GREEN)[OK]$(NC) Seeds done"

.PHONY: local-db-reseed
local-db-reseed: _local-wait-mysql build ## Drop and recreate MySQL database, then re-run migrations and seeds (services must be running)
	@echo "$(YELLOW)[WARN]$(NC) Dropping and recreating database $(PROJECT_NAME)..."
	@docker exec $(PROJECT_NAME)-mysql mysql -u root -proot \
		-e "DROP DATABASE IF EXISTS $(PROJECT_NAME); CREATE DATABASE $(PROJECT_NAME);" 2>/dev/null
	@echo "$(BLUE)[INFO]$(NC) Running migrations..."
	./$(BINARY_NAME) -c $(CONFIG_FILE) -m db_migration
	@echo "$(BLUE)[INFO]$(NC) Running seeds..."
	./$(BINARY_NAME) -c $(CONFIG_FILE) -m db_seed
	@echo "$(GREEN)[OK]$(NC) Reseed complete"

# ─── Test ────────────────────────────────────────────────────────────────────
.PHONY: local-test
local-test: ## Run unit tests (verbose, no cache; excludes e2e)
	go test -v -count=1 ./tests/units/...

.PHONY: local-test-e2e
local-test-e2e: build ## Run E2E tests (services must be up; starts app server automatically)
	@echo "$(BLUE)[INFO]$(NC) Running E2E tests..."
	@echo "$(YELLOW)[NOTE]$(NC) Ensure $(CONFIG_FILE) exists (run: make local-env-setup)"
	SCAFFOLD_E2E_BINARY=./$(BINARY_NAME) go test -v -count=1 -timeout=120s ./tests/e2e/...
	@echo "$(GREEN)[OK]$(NC) E2E tests done"

# ─── Go Toolchain ────────────────────────────────────────────────────────────
.PHONY: fmt
fmt: ## Format Go source files
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: tidy
tidy: ## Tidy go.mod / go.sum
	go mod tidy

# ─── Scaffold New Project ────────────────────────────────────────────────────
.PHONY: scaffold
scaffold: ## Create a new project from this scaffold (interactive)
	@command -v rsync >/dev/null 2>&1 || { echo "$(RED)[ERR]$(NC) rsync not found"; exit 1; }
	@read -p "New project name (e.g. my-app): " PROJ; \
	read -p "Go module path (e.g. github.com/org/my-app): " MOD; \
	read -p "Target directory (absolute path, must not exist): " DIR; \
	[ -z "$$PROJ" ] && { echo "$(RED)[ERR]$(NC) Project name required"; exit 1; }; \
	[ -z "$$MOD" ]  && { echo "$(RED)[ERR]$(NC) Module path required";  exit 1; }; \
	[ -z "$$DIR" ]  && { echo "$(RED)[ERR]$(NC) Target dir required";    exit 1; }; \
	[ -d "$$DIR" ]  && { echo "$(RED)[ERR]$(NC) Target dir already exists: $$DIR"; exit 1; }; \
	echo "$(BLUE)[INFO]$(NC) Copying scaffold to $$DIR..."; \
	rsync -a --exclude='.git' --exclude='*.darwin' --exclude='*.amd64' \
	         --exclude='evaluate/_run' --exclude='.sessions' --exclude='coverage.*' \
	         ./ "$$DIR/"; \
	echo "$(BLUE)[INFO]$(NC) Replacing module name..."; \
	LC_ALL=C find "$$DIR" -type f \( -name '*.go' -o -name 'go.mod' -o -name 'Makefile' \
	         -o -name '*.md' -o -name '*.yaml' -o -name '*.yml' \) \
	  -exec sed -i.bak "s|github.com/yetiz-org/goth-scaffold|$$MOD|g" {} +; \
	LC_ALL=C find "$$DIR" -type f \( -name '*.go' -o -name 'Makefile' -o -name 'docker-compose*.yml' \) \
	  -exec sed -i.bak "s|goth-scaffold|$$PROJ|g" {} +; \
	find "$$DIR" -name '*.bak' -delete; \
	echo "$(BLUE)[INFO]$(NC) Creating AI agent symlinks..."; \
	ln -s .agents "$$DIR/.claude"; \
	ln -s .agents "$$DIR/.codex"; \
	ln -s .agents "$$DIR/.gemini"; \
	cd "$$DIR" && git init -q && echo "$(GREEN)[OK]$(NC) New project ready at $$DIR"; \
	echo "  Next: cd $$DIR && make local-env-setup && make local-env-start && make local-db-seed && make local-run"
