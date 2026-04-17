SHELL := /bin/bash

PROJECT_NAME := scaffold
PROJECT_DB   := $(shell echo $(PROJECT_NAME) | tr '-' '_')
EVALUATE_DIR := evaluate
COMPOSE_FILE := $(EVALUATE_DIR)/docker-compose.yml
COMPOSE      ?= docker compose
CONFIG_FILE  ?= $(EVALUATE_DIR)/config.yaml.local

# Database adapter selection — drives template rendering, docker-compose
# profile, CLI tools, and reseed behaviour. Override per-invocation, e.g.
#   DB_ADAPTER=postgres make local-db-seed
DB_ADAPTER   ?= mysql
COMPOSE_BASE := $(COMPOSE) -f $(COMPOSE_FILE) --profile $(DB_ADAPTER)

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
local-env-setup: ## Generate config.yaml.local and env credentials from templates (idempotent; honours DB_ADAPTER)
	@DB_ADAPTER=$(DB_ADAPTER) bash $(EVALUATE_DIR)/scripts/generate-configs.sh

.PHONY: local-env-start
local-env-start: ## Start Docker services for the active DB_ADAPTER profile
	@echo "$(BLUE)[INFO]$(NC) Starting services (adapter=$(DB_ADAPTER))..."
	$(COMPOSE_BASE) up -d
	@echo "$(GREEN)[OK]$(NC) Services started"

.PHONY: local-env-status
local-env-status: ## Show Docker service status
	$(COMPOSE_BASE) ps

.PHONY: local-env-logs
local-env-logs: ## Follow Docker service logs (Ctrl-C to exit)
	$(COMPOSE_BASE) logs -f

.PHONY: local-env-stop
local-env-stop: ## Stop Docker services for the active DB_ADAPTER profile
	@echo "$(BLUE)[INFO]$(NC) Stopping services (adapter=$(DB_ADAPTER))..."
	$(COMPOSE_BASE) down
	@echo "$(GREEN)[OK]$(NC) Services stopped"

.PHONY: local-env-clean
local-env-clean: ## Destroy all local data volumes and config (destructive; both adapters)
	@echo "$(YELLOW)[WARN]$(NC) Destroying all local data (MySQL / Postgres / Redis / Cassandra)..."
	$(COMPOSE) -f $(COMPOSE_FILE) --profile mysql --profile postgres down -v
	@if [ -d "$(EVALUATE_DIR)/_run" ]; then \
		docker run --user root --rm -v "$$(pwd)/$(EVALUATE_DIR)/_run:/data" redis:7-alpine \
			sh -c "rm -rf /data/*" 2>/dev/null || true; \
		rm -rf $(EVALUATE_DIR)/_run; \
	fi
	@rm -rf $(EVALUATE_DIR)/env
	@echo "$(GREEN)[OK]$(NC) Environment cleaned"

# ─── Service Connections ──────────────────────────────────────────────────────
.PHONY: local-database-connect
local-database-connect: ## Open database CLI for the active DB_ADAPTER
ifeq ($(DB_ADAPTER),postgres)
	docker exec -it $(PROJECT_NAME)-postgres psql -U postgres -d $(PROJECT_DB)
else
	docker exec -it $(PROJECT_NAME)-mysql mysql -u root -proot
endif

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
local-run: local-env-setup local-env-start _local-wait-database _local-wait-cassandra _local-init-cassandra build ## Build and start API on :8080 (default mode)
	@echo "$(BLUE)[INFO]$(NC) Starting API on :8080 (mode=default, adapter=$(DB_ADAPTER))"
	./$(BINARY_NAME) -c $(CONFIG_FILE) -m default

# ─── Service Health Wait ─────────────────────────────────────────────────────
# Internal target: polls until the active database (MySQL or PostgreSQL) is
# accepting connections (max 120 s). Dispatches on DB_ADAPTER.
.PHONY: _local-wait-database
_local-wait-database:
ifeq ($(DB_ADAPTER),postgres)
	@$(MAKE) --no-print-directory _local-wait-postgres
else
	@$(MAKE) --no-print-directory _local-wait-mysql
endif

.PHONY: _local-wait-mysql
_local-wait-mysql:
	@echo "$(BLUE)[INFO]$(NC) Waiting for MySQL..."
	@elapsed=0; \
	for i in $$(seq 1 60); do \
		printf "\r$(BLUE)[INFO]$(NC) MySQL check... %3ds / 120s  " $$elapsed; \
		docker exec $(PROJECT_NAME)-mysql mysqladmin ping -h localhost --silent >/dev/null 2>&1 && \
			printf "\n" && echo "$(GREEN)[OK]$(NC) MySQL ready ($$elapsed s)" && exit 0; \
		sleep 2; \
		elapsed=$$((elapsed + 2)); \
	done; \
	printf "\n"; \
	echo "$(RED)[ERR]$(NC) MySQL not ready after 120 s" && exit 1

.PHONY: _local-wait-postgres
_local-wait-postgres:
	@echo "$(BLUE)[INFO]$(NC) Waiting for PostgreSQL..."
	@elapsed=0; \
	for i in $$(seq 1 60); do \
		printf "\r$(BLUE)[INFO]$(NC) PostgreSQL check... %3ds / 120s  " $$elapsed; \
		docker exec $(PROJECT_NAME)-postgres pg_isready -U postgres >/dev/null 2>&1 && \
			printf "\n" && echo "$(GREEN)[OK]$(NC) PostgreSQL ready ($$elapsed s)" && exit 0; \
		sleep 2; \
		elapsed=$$((elapsed + 2)); \
	done; \
	printf "\n"; \
	echo "$(RED)[ERR]$(NC) PostgreSQL not ready after 120 s" && exit 1

# Internal target: polls until Cassandra is accepting CQL connections (max 180 s).
.PHONY: _local-wait-cassandra
_local-wait-cassandra:
	@echo "$(BLUE)[INFO]$(NC) Waiting for Cassandra..."
	@elapsed=0; \
	for i in $$(seq 1 90); do \
		printf "\r$(BLUE)[INFO]$(NC) Cassandra check... %3ds / 180s  " $$elapsed; \
		docker exec $(PROJECT_NAME)-cassandra cqlsh -e "SELECT release_version FROM system.local" >/dev/null 2>&1 && \
			printf "\n" && echo "$(GREEN)[OK]$(NC) Cassandra ready ($$elapsed s)" && exit 0; \
		sleep 2; \
		elapsed=$$((elapsed + 2)); \
	done; \
	printf "\n"; \
	echo "$(RED)[ERR]$(NC) Cassandra not ready after 180 s" && exit 1

# Internal target: creates the Cassandra keyspace for this project if it does not exist.
.PHONY: _local-init-cassandra
_local-init-cassandra:
	@echo "$(BLUE)[INFO]$(NC) Initialising Cassandra keyspace $(PROJECT_DB)..."
	@docker exec $(PROJECT_NAME)-cassandra cqlsh \
		-e "CREATE KEYSPACE IF NOT EXISTS $(PROJECT_DB) WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '1'};" && \
		echo "$(GREEN)[OK]$(NC) Cassandra keyspace $(PROJECT_DB) ready"

# ─── Database Operations ─────────────────────────────────────────────────────
# local-db-migration ensures the environment is running before migrating.
# local-db-seed is the full bootstrap: setup → start services → wait → migrate → seed.
# Running `make local-env-clean && make local-db-seed` produces a clean, seeded DB.
# Running `make local-db-reseed` drops and recreates the DB then migrates + seeds.
.PHONY: local-db-migration
local-db-migration: local-env-setup local-env-start _local-wait-database _local-wait-cassandra _local-init-cassandra build ## Ensure env is up, then run database migrations
	@echo "$(BLUE)[INFO]$(NC) Running migrations (adapter=$(DB_ADAPTER))..."
	./$(BINARY_NAME) -c $(CONFIG_FILE) -m db_migration
	@echo "$(GREEN)[OK]$(NC) Migrations done"

.PHONY: local-db-seed
local-db-seed: local-db-migration ## Run migrations then seed data
	@echo "$(BLUE)[INFO]$(NC) Running seeds..."
	./$(BINARY_NAME) -c $(CONFIG_FILE) -m db_seed
	@echo "$(GREEN)[OK]$(NC) Seeds done"

.PHONY: local-db-reseed
local-db-reseed: _local-wait-database _local-wait-cassandra _local-init-cassandra build ## Drop and recreate the active database, then re-run migrations and seeds
	@echo "$(YELLOW)[WARN]$(NC) Dropping and recreating database $(PROJECT_DB) (adapter=$(DB_ADAPTER))..."
ifeq ($(DB_ADAPTER),postgres)
	@docker exec $(PROJECT_NAME)-postgres psql -U postgres -d postgres \
		-c "DROP DATABASE IF EXISTS $(PROJECT_DB); CREATE DATABASE $(PROJECT_DB);" 2>/dev/null
else
	@docker exec $(PROJECT_NAME)-mysql mysql -u root -proot \
		-e "DROP DATABASE IF EXISTS $(PROJECT_DB); CREATE DATABASE $(PROJECT_DB);" 2>/dev/null
endif
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
	PROJ=$$(echo "$$PROJ" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]/-/g; s/-\{2,\}/-/g; s/^-//; s/-$$//'); \
	[ -z "$$PROJ" ] && { echo "$(RED)[ERR]$(NC) Invalid project name after sanitization"; exit 1; }; \
	PROJ_DB=$$(echo "$$PROJ" | tr '-' '_'); \
	PROJ_UPPER=$$(echo "$$PROJ" | tr '[:lower:]' '[:upper:]' | tr '-' '_'); \
	[ -d "$$DIR" ]  && { echo "$(RED)[ERR]$(NC) Target dir already exists: $$DIR"; exit 1; }; \
	echo "$(BLUE)[INFO]$(NC) Copying scaffold to $$DIR..."; \
	rsync -a --filter=':- .gitignore' \
	         --exclude='.git' --exclude='alloc/' \
	         --exclude='LICENSE' --exclude='LICENSE.KKLAB' --exclude='NOTICE' --exclude='README.md' \
	         ./ "$$DIR/"; \
	cp ./evaluate/templates/config.yaml.template "$$DIR/evaluate/templates/config.yaml.template"; \
	echo "$(BLUE)[INFO]$(NC) Removing scaffold target from new project..."; \
	sed -i.bak '/^# ─── Scaffold New Project/,$$d' "$$DIR/Makefile"; \
	echo "$(BLUE)[INFO]$(NC) Replacing module name..."; \
	LC_ALL=C find "$$DIR" -type f \( -name '*.go' -o -name 'go.mod' -o -name 'Makefile' \
	         -o -name '*.md' -o -name '*.yaml' -o -name '*.yml' -o -name '*.html' \) \
	  -exec sed -i.bak "s|github.com/yetiz-org/goth-scaffold|$$MOD|g" {} +; \
	LC_ALL=C find "$$DIR" -type f \( -name '*.go' -o -name 'Makefile' -o -name '*.md' \
	         -o -name '*.yaml' -o -name '*.yml' -o -name '*.env' -o -name '*.html' \) \
	  -exec sed -i.bak "s|goth-scaffold|$$PROJ|g; s|Goth Scaffold|$$PROJ|g; s|Goth-Scaffold|$$PROJ|g" {} +; \
	sed -i.bak "s|^PROJECT_NAME := scaffold$$|PROJECT_NAME := $$PROJ|" "$$DIR/Makefile"; \
	echo "$(BLUE)[INFO]$(NC) Replacing remaining scaffold identifiers..."; \
	LC_ALL=C find "$$DIR" -type f \( -name '*.go' -o -name 'Makefile' -o -name '*.md' \
	         -o -name '*.yaml' -o -name '*.yml' -o -name '*.env' -o -name '.gitignore' \
	         -o -name '*.html' -o -name '*.json' -o -name '*.template' -o -name '*.sh' \) \
	  -exec sed -i.bak "s|SCAFFOLD|$$PROJ_UPPER|g; s|Scaffold|$$PROJ|g; s|scaffold|$$PROJ|g" {} +; \
	echo "$(BLUE)[INFO]$(NC) Fixing datastore names for database compatibility (hyphen → underscore)..."; \
	LC_ALL=C find "$$DIR" -name 'config.yaml*' \
	  -exec sed -i.bak \
	    "s|database_name: \"$$PROJ\"|database_name: \"$$PROJ_DB\"|g; \
	     s|cassandra_name: \"$$PROJ\"|cassandra_name: \"$$PROJ_DB\"|g; \
	     s|redis_name: \"$$PROJ\"|redis_name: \"$$PROJ_DB\"|g" {} +; \
	LC_ALL=C find "$$DIR/$(EVALUATE_DIR)/templates" -name 'defaults.env' \
	  -exec sed -i.bak \
	    "s|MYSQL_DATABASE=$$PROJ|MYSQL_DATABASE=$$PROJ_DB|; \
	     s|DB_NAME=$$PROJ|DB_NAME=$$PROJ_DB|; \
	     s|DB_NAME_YAML=$$PROJ|DB_NAME_YAML=$$PROJ_DB|; \
	     s|CASSANDRA_KEYSPACE=$$PROJ|CASSANDRA_KEYSPACE=$$PROJ_DB|; \
	     s|CASSANDRA_NAME_YAML=$$PROJ|CASSANDRA_NAME_YAML=$$PROJ_DB|; \
	     s|REDIS_NAME_YAML=$$PROJ|REDIS_NAME_YAML=$$PROJ_DB|" {} +; \
	find "$$DIR" -name '*.bak' -delete; \
	echo "$(BLUE)[INFO]$(NC) Formatting Go source files..."; \
	cd "$$DIR" && go fmt ./... && git init -q && echo "$(GREEN)[OK]$(NC) New project ready at $$DIR"; \
	echo "  Next: cd $$DIR && make local-run"
