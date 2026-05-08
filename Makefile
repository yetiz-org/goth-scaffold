SHELL := /bin/bash

PROJECT_NAME := scaffold
PROJECT_DB   := $(shell echo $(PROJECT_NAME) | tr '-' '_')
EVALUATE_DIR := evaluate
COMPOSE_FILE := $(EVALUATE_DIR)/docker-compose.yml
COMPOSE      ?= docker compose

# Database adapter selection — drives template rendering, docker-compose
# profile, CLI tools, and reseed behaviour. Override per-invocation, e.g.
#   DB_ADAPTER=postgres make local-db-seed
DB_ADAPTER   ?= mysql

COMPOSE_SCOPE       ?= local
ENV_DIR             ?= $(EVALUATE_DIR)/env/$(COMPOSE_SCOPE)
RUN_DIR             ?= $(EVALUATE_DIR)/_run/$(COMPOSE_SCOPE)
CONFIG_FILE         ?= $(RUN_DIR)/config.yaml.local
COMPOSE_RUN_DIR     ?= ./_run/$(COMPOSE_SCOPE)
COMPOSE_PROJECT     ?= $(PROJECT_NAME)-$(subst /,-,$(COMPOSE_SCOPE))
CONTAINER_PREFIX    ?= $(COMPOSE_PROJECT)
MYSQL_PORT          ?= 3306
POSTGRES_PORT       ?= 5432
CASSANDRA_PORT      ?= 9042
REDIS_PORT          ?= 6379
ASYNQMON_PORT       ?= 8081
APP_PORT            ?= 8080

ifeq ($(DB_ADAPTER),postgres)
DB_PORT_FOR_CONFIG := $(POSTGRES_PORT)
else
DB_PORT_FOR_CONFIG := $(MYSQL_PORT)
endif

COMPOSE_ENV  := COMPOSE_PROJECT_NAME=$(COMPOSE_PROJECT) CONTAINER_PREFIX=$(CONTAINER_PREFIX) RUN_DIR=$(COMPOSE_RUN_DIR) MYSQL_PORT=$(MYSQL_PORT) POSTGRES_PORT=$(POSTGRES_PORT) CASSANDRA_PORT=$(CASSANDRA_PORT) REDIS_PORT=$(REDIS_PORT) ASYNQMON_PORT=$(ASYNQMON_PORT)
COMPOSE_BASE := $(COMPOSE_ENV) $(COMPOSE) -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT) --profile $(DB_ADAPTER)

MYSQL_CONTAINER     := $(CONTAINER_PREFIX)-mysql
POSTGRES_CONTAINER  := $(CONTAINER_PREFIX)-postgres
CASSANDRA_CONTAINER := $(CONTAINER_PREFIX)-cassandra
REDIS_CONTAINER     := $(CONTAINER_PREFIX)-redis

DQ := "
BT := `

ifndef WORKTREE_ID
WORKTREE_ID := $(notdir $(CURDIR))
endif
WORKTREE_ID_RAW := $(value WORKTREE_ID)
WORKTREE_ID_SAFE := $(subst \,_bs_,$(subst $$,_dol_,$(subst $(BT),_bt_,$(subst $(DQ),_dq_,$(value WORKTREE_ID_RAW)))))
WORKTREE_ID_HASH := $(shell printf '%s' "$(WORKTREE_ID_SAFE)" | cksum | awk '{print $$1}')
WORKTREE_ID_SLUG := $(shell printf '%s' "$(WORKTREE_ID_SAFE)" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9_.-]/-/g; s/-\{2,\}/-/g; s/^-//; s/-$$//')
WORKTREE_PATH_ID := $(if $(strip $(WORKTREE_ID_SLUG)),$(WORKTREE_ID_SLUG),worktree)-$(WORKTREE_ID_HASH)
WORKTREE_COMPOSE_ID := $(shell printf '%s' "$(WORKTREE_PATH_ID)" | sed 's/[^a-z0-9_-]/-/g; s/-\{2,\}/-/g; s/^-//; s/-$$//')
WORKTREE_PORT_BASE ?= $(shell printf '%s' "$(WORKTREE_ID_SAFE)" | cksum | awk '{print 20000 + ($$1 % 20000)}')
WORKTREE_MYSQL_PORT     ?= $(WORKTREE_PORT_BASE)
WORKTREE_POSTGRES_PORT  ?= $(shell expr $(WORKTREE_PORT_BASE) + 1)
WORKTREE_CASSANDRA_PORT ?= $(shell expr $(WORKTREE_PORT_BASE) + 2)
WORKTREE_REDIS_PORT     ?= $(shell expr $(WORKTREE_PORT_BASE) + 3)
WORKTREE_ASYNQMON_PORT  ?= $(shell expr $(WORKTREE_PORT_BASE) + 4)
WORKTREE_APP_PORT       ?= $(shell expr $(WORKTREE_PORT_BASE) + 5)
WORKTREE_SCOPE          := worktree/$(WORKTREE_PATH_ID)
WORKTREE_RUN_DIR        := $(EVALUATE_DIR)/_run/$(WORKTREE_SCOPE)
WORKTREE_PORTS_CMD      := bash $(EVALUATE_DIR)/scripts/worktree-ports.sh "$(WORKTREE_RUN_DIR)" "$(WORKTREE_ID_SAFE)" "$(WORKTREE_PORT_BASE)"
WORKTREE_VARS           := COMPOSE_SCOPE=$(WORKTREE_SCOPE) ENV_DIR=$(EVALUATE_DIR)/env/$(WORKTREE_SCOPE) RUN_DIR=$(WORKTREE_RUN_DIR) CONFIG_FILE=$(WORKTREE_RUN_DIR)/config.yaml.local COMPOSE_RUN_DIR=./_run/$(WORKTREE_SCOPE) COMPOSE_PROJECT=$(PROJECT_NAME)-worktree-$(WORKTREE_COMPOSE_ID) CONTAINER_PREFIX=$(PROJECT_NAME)-worktree-$(WORKTREE_COMPOSE_ID)

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
local-env-setup: ## Generate scoped local config and env credentials from templates (idempotent; honours DB_ADAPTER)
	@mkdir -p $(RUN_DIR)
	@DB_ADAPTER=$(DB_ADAPTER) OUTPUT_ENV_DIR=$(ENV_DIR) OUTPUT_CONFIG_FILE=$(CONFIG_FILE) SECRET_PATH=$(ENV_DIR)/ DB_PORT=$(DB_PORT_FOR_CONFIG) REDIS_PORT=$(REDIS_PORT) CASSANDRA_PORT=$(CASSANDRA_PORT) APP_PORT=$(APP_PORT) bash $(EVALUATE_DIR)/scripts/generate-configs.sh

.PHONY: local-env-start
local-env-start: ## Start scoped Docker services for the active DB_ADAPTER profile
	@echo "$(BLUE)[INFO]$(NC) Starting services (scope=$(COMPOSE_SCOPE), adapter=$(DB_ADAPTER))..."
	$(COMPOSE_BASE) up -d
	@echo "$(GREEN)[OK]$(NC) Services started"

.PHONY: local-env-status
local-env-status: ## Show scoped Docker service status
	$(COMPOSE_BASE) ps

.PHONY: local-env-logs
local-env-logs: ## Follow scoped Docker service logs (Ctrl-C to exit)
	$(COMPOSE_BASE) logs -f

.PHONY: local-env-stop
local-env-stop: ## Stop scoped Docker services for the active DB_ADAPTER profile
	@echo "$(BLUE)[INFO]$(NC) Stopping services (scope=$(COMPOSE_SCOPE), adapter=$(DB_ADAPTER))..."
	$(COMPOSE_BASE) down
	@echo "$(GREEN)[OK]$(NC) Services stopped"

.PHONY: local-env-clean
local-env-clean: ## Destroy scoped local data and config only (destructive; both adapters)
	@echo "$(YELLOW)[WARN]$(NC) Destroying scoped data (scope=$(COMPOSE_SCOPE))..."
	$(COMPOSE_ENV) $(COMPOSE) -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT) --profile mysql --profile postgres down -v --remove-orphans
	@if [ "$(COMPOSE_SCOPE)" = "local" ]; then \
		for c in $(PROJECT_NAME)-mysql $(PROJECT_NAME)-postgres $(PROJECT_NAME)-cassandra $(PROJECT_NAME)-redis $(PROJECT_NAME)-asynqmon; do \
			if docker ps -a --format '{{.Names}}' | grep -qx "$$c"; then \
				docker rm -f "$$c" >/dev/null; \
			fi; \
		done; \
		for n in $(PROJECT_NAME)-network $(PROJECT_NAME)_scaffold-network $(EVALUATE_DIR)_scaffold-network scaffold-network; do \
			if docker network ls --format '{{.Name}}' | grep -qx "$$n"; then \
				docker network rm "$$n" >/dev/null 2>&1 || true; \
			fi; \
		done; \
		for d in $(EVALUATE_DIR)/_run/mysql $(EVALUATE_DIR)/_run/postgres $(EVALUATE_DIR)/_run/cassandra $(EVALUATE_DIR)/_run/redis; do \
			if [ -d "$$d" ]; then \
				docker run --user root --rm -v "$$(pwd)/$$d:/data" redis:7-alpine \
					sh -c "rm -rf /data/*" 2>/dev/null || true; \
				rm -rf "$$d"; \
			fi; \
		done; \
		find $(EVALUATE_DIR)/env -mindepth 1 -maxdepth 1 -type d \
			\( -name 'database-*' -o -name 'redis-*' -o -name 'cassandra-*' \) \
			-exec rm -rf {} + 2>/dev/null || true; \
		rm -f $(EVALUATE_DIR)/config.yaml.local; \
	fi
	@if [ -d "$(RUN_DIR)" ]; then \
		docker run --user root --rm -v "$$(pwd)/$(RUN_DIR):/data" redis:7-alpine \
			sh -c "rm -rf /data/*" 2>/dev/null || true; \
		rm -rf $(RUN_DIR); \
	fi
	@rm -rf $(ENV_DIR)
	@env_parent="$$(dirname "$(ENV_DIR)")"; \
	while [ "$$env_parent" != "." ] && [ "$$env_parent" != "$(EVALUATE_DIR)" ]; do \
		rmdir "$$env_parent" 2>/dev/null || break; \
		env_parent="$$(dirname "$$env_parent")"; \
	done; \
	rmdir "$(EVALUATE_DIR)/env" 2>/dev/null || true
	@run_parent="$$(dirname "$(RUN_DIR)")"; \
	while [ "$$run_parent" != "." ] && [ "$$run_parent" != "$(EVALUATE_DIR)" ]; do \
		rmdir "$$run_parent" 2>/dev/null || break; \
		run_parent="$$(dirname "$$run_parent")"; \
	done; \
	rmdir "$(EVALUATE_DIR)/_run" 2>/dev/null || true
	@echo "$(GREEN)[OK]$(NC) Environment cleaned"

# ─── Service Connections ──────────────────────────────────────────────────────
.PHONY: local-database-connect
local-database-connect: ## Open database CLI for the active DB_ADAPTER
ifeq ($(DB_ADAPTER),postgres)
	docker exec -it $(POSTGRES_CONTAINER) psql -U postgres -d $(PROJECT_DB)
else
	docker exec -it $(MYSQL_CONTAINER) mysql -u root -proot
endif

.PHONY: local-redis-connect
local-redis-connect: ## Open Redis CLI
	docker exec -it $(REDIS_CONTAINER) redis-cli

.PHONY: local-cassandra-connect
local-cassandra-connect: ## Open Cassandra CQL shell
	docker exec -it $(CASSANDRA_CONTAINER) cqlsh

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
local-run: local-env-setup local-env-start _local-wait-database _local-wait-cassandra _local-init-cassandra build ## Build and start API on the scoped APP_PORT (default mode)
	@echo "$(BLUE)[INFO]$(NC) Starting API on :$(APP_PORT) (scope=$(COMPOSE_SCOPE), mode=default, adapter=$(DB_ADAPTER))"
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
		docker exec $(MYSQL_CONTAINER) mysqladmin ping -h localhost --silent >/dev/null 2>&1 && \
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
		docker exec $(POSTGRES_CONTAINER) pg_isready -U postgres >/dev/null 2>&1 && \
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
		docker exec $(CASSANDRA_CONTAINER) cqlsh -e "SELECT release_version FROM system.local" >/dev/null 2>&1 && \
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
	@docker exec $(CASSANDRA_CONTAINER) cqlsh \
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
	@docker exec $(POSTGRES_CONTAINER) psql -U postgres -d postgres \
		-c "DROP DATABASE IF EXISTS $(PROJECT_DB); CREATE DATABASE $(PROJECT_DB);" 2>/dev/null
else
	@docker exec $(MYSQL_CONTAINER) mysql -u root -proot \
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
	SCAFFOLD_E2E_BINARY=./$(BINARY_NAME) SCAFFOLD_E2E_CONFIG=$(CONFIG_FILE) go test -v -count=1 -timeout=120s ./tests/e2e/...
	@echo "$(GREEN)[OK]$(NC) E2E tests done"

# ─── Worktree Environment (isolated Docker Compose scope) ───────────────────
.PHONY: _worktree-guard
_worktree-guard:
	@[ -n "$(WORKTREE_PATH_ID)" ] && [ -n "$(WORKTREE_COMPOSE_ID)" ] || { \
		echo "$(RED)[ERR]$(NC) WORKTREE_ID=$(WORKTREE_ID_SAFE) does not produce a safe worktree scope"; \
		exit 1; \
	}

.PHONY: worktree-env-setup
worktree-env-setup: _worktree-guard ## Generate isolated env/worktree/<safe-id> and _run/worktree/<safe-id>/config.yaml.local
	@ports="$$($(WORKTREE_PORTS_CMD))" || exit 1; $(MAKE) --no-print-directory $(WORKTREE_VARS) $$ports local-env-setup

.PHONY: worktree-env-start
worktree-env-start: _worktree-guard ## Start isolated Docker services for this worktree
	@ports="$$($(WORKTREE_PORTS_CMD))" || exit 1; $(MAKE) --no-print-directory $(WORKTREE_VARS) $$ports local-env-start

.PHONY: worktree-env-status
worktree-env-status: _worktree-guard ## Show isolated Docker service status for this worktree
	@ports="$$($(WORKTREE_PORTS_CMD))" || exit 1; $(MAKE) --no-print-directory $(WORKTREE_VARS) $$ports local-env-status

.PHONY: worktree-env-logs
worktree-env-logs: _worktree-guard ## Follow isolated Docker service logs for this worktree
	@ports="$$($(WORKTREE_PORTS_CMD))" || exit 1; $(MAKE) --no-print-directory $(WORKTREE_VARS) $$ports local-env-logs

.PHONY: worktree-env-stop
worktree-env-stop: _worktree-guard ## Stop isolated Docker services for this worktree
	@ports="$$($(WORKTREE_PORTS_CMD))" || exit 1; $(MAKE) --no-print-directory $(WORKTREE_VARS) $$ports local-env-stop

.PHONY: worktree-env-clean
worktree-env-clean: _worktree-guard ## Destroy only this worktree's containers, env, and _run data
	@ports="$$($(WORKTREE_PORTS_CMD))" || exit 1; $(MAKE) --no-print-directory $(WORKTREE_VARS) $$ports local-env-clean

.PHONY: worktree-run
worktree-run: _worktree-guard ## Build and start this worktree API on its isolated APP_PORT
	@ports="$$($(WORKTREE_PORTS_CMD))" || exit 1; $(MAKE) --no-print-directory $(WORKTREE_VARS) $$ports local-run

.PHONY: worktree-db-migration
worktree-db-migration: _worktree-guard ## Run migrations in this worktree's isolated environment
	@ports="$$($(WORKTREE_PORTS_CMD))" || exit 1; $(MAKE) --no-print-directory $(WORKTREE_VARS) $$ports local-db-migration

.PHONY: worktree-db-seed
worktree-db-seed: _worktree-guard ## Run migrations then seeds in this worktree's isolated environment
	@ports="$$($(WORKTREE_PORTS_CMD))" || exit 1; $(MAKE) --no-print-directory $(WORKTREE_VARS) $$ports local-db-seed

.PHONY: worktree-db-reseed
worktree-db-reseed: _worktree-guard ## Drop/recreate this worktree's active database, then migrate and seed
	@ports="$$($(WORKTREE_PORTS_CMD))" || exit 1; $(MAKE) --no-print-directory $(WORKTREE_VARS) $$ports local-db-reseed

.PHONY: worktree-test-e2e
worktree-test-e2e: _worktree-guard ## Run E2E tests against this worktree's isolated config
	@ports="$$($(WORKTREE_PORTS_CMD))" || exit 1; $(MAKE) --no-print-directory $(WORKTREE_VARS) $$ports local-test-e2e

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
