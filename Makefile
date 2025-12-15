SHELL := /bin/bash

EVALUATE_DIR := evaluate
COMPOSE_FILE := $(EVALUATE_DIR)/docker-compose.yml
COMPOSE ?= docker compose

.PHONY: env-up
env-up:
	$(COMPOSE) -f $(COMPOSE_FILE) up -d

.PHONY: env-down
env-down:
	$(COMPOSE) -f $(COMPOSE_FILE) down

.PHONY: env-ps
env-ps:
	$(COMPOSE) -f $(COMPOSE_FILE) ps

.PHONY: env-logs
env-logs:
	$(COMPOSE) -f $(COMPOSE_FILE) logs -f

.PHONY: env-restart
env-restart: env-down env-up

.PHONY: test
test:
	go test -v ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: build
build:
	go build ./...
