GO ?= go
GOLANGCI_LINT ?= golangci-lint
BINARY := agentflow
CMD := ./application/cmd/agentflow
BIN_DIR := bin
RELEASE_VERSION ?= dev
LDFLAGS := -ldflags "-X github.com/LaProgrammerie/hyper-fast-builder/application/internal/version.Version=$(RELEASE_VERSION)"
COMPOSE := docker compose -f infrastructure/docker/docker-compose.yml

.PHONY: help build run test lint fmt vet clean dev docker-up docker-down benchmark

benchmark: build ## Benchmark dry-run (estimate + work plan-only)
	@AGENTFLOW_DRY_RUN=1 ./scripts/benchmark-workflow.sh

help: ## Affiche les cibles disponibles
	@grep -E '^[a-zA-Z0-9_.-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  %-16s %s\n", $$1, $$2}'

build: ## Compile le binaire dans bin/
	@mkdir -p $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY) $(CMD)

run: build ## Lance agentflow (ARGS=...)
	$(BIN_DIR)/$(BINARY) $(ARGS)

test: ## Lance les tests
	$(GO) test ./...

test-cover: ## Tests avec couverture
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

lint: ## Analyse statique (golangci-lint)
	$(GOLANGCI_LINT) run

fmt: ## Formate le code
	$(GO) fmt ./...

vet: ## go vet
	$(GO) vet ./...

clean: ## Supprime les binaires locaux
	rm -rf $(BIN_DIR) coverage.out

dev: docker-up ## Environnement de dev (conteneur Go + deps)

docker-up: ## Démarre le service dev Docker
	$(COMPOSE) up -d --build

docker-down: ## Arrête la stack Docker
	$(COMPOSE) down

docker-shell: ## Shell dans le conteneur dev
	$(COMPOSE) exec dev bash
