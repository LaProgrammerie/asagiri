GO ?= go
GOLANGCI_LINT ?= golangci-lint
BINARY := asa
CMD := ./application/cmd/asa
BIN_DIR := bin
VERSION_PKG := github.com/LaProgrammerie/asagiri/application/internal/version
RELEASE_VERSION ?= dev
RELEASE_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
RELEASE_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo unknown)
RELEASE_LDFLAGS := -s -w \
	-X $(VERSION_PKG).Version=$(RELEASE_VERSION) \
	-X $(VERSION_PKG).Commit=$(RELEASE_COMMIT) \
	-X $(VERSION_PKG).Date=$(RELEASE_DATE)
LDFLAGS := -ldflags "$(RELEASE_LDFLAGS)"
COMPOSE := docker compose -f infrastructure/docker/docker-compose.yml

.PHONY: help build run test lint fmt vet clean dev docker-up docker-down benchmark release-snapshot release-check

benchmark: build ## Benchmark dry-run (estimate + work plan-only)
	@ASA_DRY_RUN=1 ./scripts/benchmark-workflow.sh

help: ## Affiche les cibles disponibles
	@grep -E '^[a-zA-Z0-9_.-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  %-16s %s\n", $$1, $$2}'

build: ## Compile le binaire dans bin/
	@mkdir -p $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY) $(CMD)

run: build ## Lance asa (ARGS=...)
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
	rm -rf $(BIN_DIR) coverage.out dist/

release-snapshot: ## Build release artefacts locally (GoReleaser snapshot)
	goreleaser release --snapshot --clean

release-check: ## Validate .goreleaser.yaml
	goreleaser check

dev: docker-up ## Environnement de dev (conteneur Go + deps)

docker-up: ## Démarre le service dev Docker
	$(COMPOSE) up -d --build

docker-down: ## Arrête la stack Docker
	$(COMPOSE) down

docker-shell: ## Shell dans le conteneur dev
	$(COMPOSE) exec dev bash
