# Makefile for luxe-backend (PostgreSQL + Goose)
#
# Defaults target an existing local Postgres container (docker-psql_bp-1 on :5432).
# Override any value via environment, .env, or on the command line:
#   make migrate-up POSTGRES_CONTAINER=my-postgres POSTGRES_DB=luxe
#   make dev-setup USE_COMPOSE=1   # start luxe-backend compose stack instead

# Load local overrides when present (.env is gitignored)
-include .env
export

# ─── App ──────────────────────────────────────────────────────────────────────
BINARY_NAME ?= luxe-api
GOBASE      := $(shell pwd)
GOBIN       := $(GOBASE)/bin
MIGRATIONS_DIR := ./internal/migrations

# ─── Postgres (existing container / local DB) ─────────────────────────────────
# If DATABASE_URL is set in .env, it wins. Otherwise we build from the fields below.
POSTGRES_CONTAINER ?= docker-psql_bp-1

# Auto-detect user/db/password from a running container when not set in .env
DETECTED_DB_USER     := $(shell docker exec $(POSTGRES_CONTAINER) printenv POSTGRES_USER 2>/dev/null)
DETECTED_DB_NAME     := $(shell docker exec $(POSTGRES_CONTAINER) printenv POSTGRES_DB 2>/dev/null)
DETECTED_DB_PASSWORD := $(shell docker exec $(POSTGRES_CONTAINER) printenv POSTGRES_PASSWORD 2>/dev/null)

POSTGRES_HOST      ?= $(or $(DB_HOST),localhost)
POSTGRES_PORT      ?= $(or $(DB_PORT),5432)
POSTGRES_USER      ?= $(or $(DB_USER),$(DETECTED_DB_USER),postgres)
POSTGRES_PASSWORD  ?= $(or $(DB_PASSWORD),$(DETECTED_DB_PASSWORD),postgres)
POSTGRES_DB        ?= $(or $(DB_NAME),shopping_platform)
POSTGRES_SSLMODE   ?= $(or $(DB_SSLMODE),disable)

# Reject accidental full-URL values in DB_HOST (common .env mistake).
ifeq ($(findstring ://,$(POSTGRES_HOST)),://)
POSTGRES_HOST := localhost
endif

ifndef DATABASE_URL
DATABASE_URL := postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)
endif

# ─── Docker Compose (optional — set USE_COMPOSE=1 for project-managed stack) ───
USE_COMPOSE          ?= 0
COMPOSE              ?= docker compose
COMPOSE_FILE         ?= docker-compose.yml
COMPOSE_PROJECT_NAME ?= luxe-backend
POSTGRES_SERVICE     ?= postgres
REDIS_SERVICE        ?= redis
JAEGER_SERVICE       ?= jaeger

# In-compose DSN (service hostname `postgres`, internal port 5432)
GOOSE_DOCKER_URL ?= postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_SERVICE):5432/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)

GOOSE_CMD = goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)"

# Colors for output
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: help build run clean test test-coverage lint migrate-create migrate-up migrate-up-docker migrate-down migrate-reset migrate-status migrate-force deps tidy install-tools docker-up docker-up-jaeger docker-wait-postgres stripe-listen dev-setup dev-setup-existing seed-dev seed-shipping-providers seed-invoices seed-coupons seed-nav-menus-i18n seed-catalog-i18n seed-orders-returns db-info

# Default target
help: ## Show this help message
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Database overrides (env or command line):'
	@echo '  POSTGRES_CONTAINER=$(POSTGRES_CONTAINER)'
	@echo '  DATABASE_URL=$(DATABASE_URL)'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${YELLOW}%-22s${RESET} ${GREEN}%s${RESET}\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

db-info: ## Print resolved database / container settings
	@echo "${GREEN}Postgres container:${RESET} $(POSTGRES_CONTAINER)"
	@echo "${GREEN}Postgres user:${RESET}      $(POSTGRES_USER)"
	@echo "${GREEN}Postgres database:${RESET}  $(POSTGRES_DB)"
	@echo "${GREEN}DATABASE_URL:${RESET}         postgresql://$(POSTGRES_USER):***@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)"
	@echo "${GREEN}USE_COMPOSE:${RESET}          $(USE_COMPOSE)"
	@if [ -z "$(DB_PASSWORD)" ] && [ -n "$(DETECTED_DB_PASSWORD)" ]; then \
		echo "${GREEN}Password:${RESET}             loaded from container env"; \
	elif [ "$(POSTGRES_PASSWORD)" = "postgres" ] && [ "$(POSTGRES_USER)" != "postgres" ]; then \
		echo "${YELLOW}Hint: set DB_PASSWORD in .env or ensure $(POSTGRES_CONTAINER) is running.${RESET}"; \
	fi

db-detect: ## Show Postgres user/db from the running container
	@echo "${GREEN}Container:${RESET} $(POSTGRES_CONTAINER)"
	@echo "${GREEN}POSTGRES_USER:${RESET} $(DETECTED_DB_USER)"
	@echo "${GREEN}POSTGRES_DB (container):${RESET} $(DETECTED_DB_NAME) ${YELLOW}(ecom-go — do not migrate luxe onto this)${RESET}"
	@if [ -n "$(DETECTED_DB_PASSWORD)" ]; then \
		echo "${GREEN}POSTGRES_PASSWORD:${RESET} set in container (auto-used when DB_PASSWORD is empty)"; \
	else \
		echo "${YELLOW}POSTGRES_PASSWORD:${RESET} not found in container env"; \
	fi
	@echo "${YELLOW}Optional .env overrides:${RESET}"
	@echo "  DB_USER=$(DETECTED_DB_USER)"
	@echo "  DB_NAME=$(DETECTED_DB_NAME)"
	@echo "  DB_PORT=5432"
	@echo "  DB_PASSWORD=...  # only if you want to override container value"

db-create: ## Create POSTGRES_DB on the running container (for luxe-backend; do not use ecom-go's ecom db)
	@echo "${GREEN}Creating database '$(POSTGRES_DB)' on $(POSTGRES_CONTAINER)...${RESET}"
	@docker exec $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d postgres -tc \
		"SELECT 1 FROM pg_database WHERE datname = '$(POSTGRES_DB)'" | grep -q 1 \
		&& echo "${YELLOW}Database '$(POSTGRES_DB)' already exists${RESET}" \
		|| docker exec $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d postgres -c \
		"CREATE DATABASE \"$(POSTGRES_DB)\" OWNER \"$(POSTGRES_USER)\";"
	@echo "${GREEN}Done. Set DB_NAME=$(POSTGRES_DB) in .env then: make migrate-up${RESET}"

## Development
build: ## Build the application
	@echo "${GREEN}Building application...${RESET}"
	go build -o $(GOBIN)/$(BINARY_NAME) ./cmd/api
	@echo "${GREEN}Build complete: $(GOBIN)/$(BINARY_NAME)${RESET}"

run: ## Run the application (uses .env or environment variables)
	@echo "${GREEN}Running application...${RESET}"
	go run ./cmd/api

docker-up: ## Start Postgres and Redis via compose (USE_COMPOSE=1) or verify existing container
	@if [ "$(USE_COMPOSE)" = "1" ]; then \
		echo "${GREEN}Starting compose stack ($(COMPOSE_PROJECT_NAME))...${RESET}"; \
		COMPOSE_PROJECT_NAME=$(COMPOSE_PROJECT_NAME) $(COMPOSE) -f $(COMPOSE_FILE) up -d $(POSTGRES_SERVICE) $(REDIS_SERVICE); \
	else \
		echo "${GREEN}Using existing Postgres container: $(POSTGRES_CONTAINER)${RESET}"; \
		docker ps --format '{{.Names}}' | grep -qx '$(POSTGRES_CONTAINER)' || { \
			echo "${YELLOW}Container '$(POSTGRES_CONTAINER)' is not running.${RESET}"; \
			echo "${YELLOW}Start it, or run: make docker-up USE_COMPOSE=1${RESET}"; \
			exit 1; \
		}; \
	fi
	@echo "${GREEN}Postgres ready target: $(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)${RESET}"

docker-wait-postgres: ## Wait until Postgres accepts connections
	@echo "${GREEN}Waiting for Postgres ($(POSTGRES_CONTAINER) @ $(POSTGRES_HOST):$(POSTGRES_PORT))...${RESET}"
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30; do \
		if [ "$(USE_COMPOSE)" = "1" ]; then \
			COMPOSE_PROJECT_NAME=$(COMPOSE_PROJECT_NAME) $(COMPOSE) -f $(COMPOSE_FILE) exec -T $(POSTGRES_SERVICE) pg_isready -U $(POSTGRES_USER) >/dev/null 2>&1 && { echo "${GREEN}Postgres ready (compose)${RESET}"; exit 0; }; \
		else \
			docker exec $(POSTGRES_CONTAINER) pg_isready -U $(POSTGRES_USER) >/dev/null 2>&1 && { echo "${GREEN}Postgres ready ($(POSTGRES_CONTAINER))${RESET}"; exit 0; }; \
		fi; \
		sleep 1; \
	done; \
	echo "${YELLOW}Postgres not ready after 30s — check: make db-info${RESET}"; exit 1

docker-up-jaeger: ## Start Jaeger for OTLP tracing (optional; needs Docker Hub access)
	@echo "${GREEN}Starting Jaeger...${RESET}"
	COMPOSE_PROJECT_NAME=$(COMPOSE_PROJECT_NAME) $(COMPOSE) -f $(COMPOSE_FILE) up -d $(JAEGER_SERVICE) || (echo "${YELLOW}Jaeger failed to start (network/Docker Hub?). OTEL_ENABLED still works if a collector is reachable.${RESET}" && exit 1)

docker-up-all: docker-up docker-up-jaeger ## Start Postgres, Redis, and Jaeger

stripe-listen: ## Forward Stripe webhooks (orders + wallet deposits) to local API
	@echo "${GREEN}Forwarding Stripe webhooks to localhost:8080/api/v1/webhooks/stripe${RESET}"
	stripe listen --forward-to localhost:8080/api/v1/webhooks/stripe --events checkout.session.completed,checkout.session.expired

dev-setup-existing: docker-wait-postgres migrate-up ## Wait for existing Postgres container and run migrations
	@echo "${GREEN}Dev setup complete (existing container)${RESET}"

dev-setup: ## Wait for Postgres and run migrations (compose stack when USE_COMPOSE=1)
	@$(MAKE) docker-up
	@$(MAKE) docker-wait-postgres
	@$(MAKE) migrate-up
	@echo "${GREEN}Dev setup complete${RESET}"

seed-dev: ## Load dev demo data (local/staging only; uses psql or docker exec)
	@echo "${GREEN}Seeding dev demo data into $(POSTGRES_DB) via $(POSTGRES_CONTAINER)...${RESET}"
	@if command -v psql >/dev/null 2>&1; then \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-dev.sql; \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-catalog.sql; \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-shipping-providers.sql; \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-orders-returns.sql; \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-invoices.sql; \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-coupons.sql; \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-nav-menus-i18n.sql; \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-catalog-i18n.sql; \
	else \
		echo "${YELLOW}psql not found — seeding via docker exec $(POSTGRES_CONTAINER)${RESET}"; \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-dev.sql; \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-catalog.sql; \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-shipping-providers.sql; \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-orders-returns.sql; \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-invoices.sql; \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-coupons.sql; \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-nav-menus-i18n.sql; \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-catalog-i18n.sql; \
	fi
	@echo "${GREEN}Dev seed complete${RESET}"

seed-shipping-providers: ## Load demo shipping providers (checkout + admin carriers page)
	@echo "${GREEN}Seeding shipping providers into $(POSTGRES_DB)...${RESET}"
	@if command -v psql >/dev/null 2>&1; then \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-shipping-providers.sql; \
	else \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-shipping-providers.sql; \
	fi
	@echo "${GREEN}Shipping providers seed complete${RESET}"

seed-invoices: ## Load demo invoices from seed orders
	@echo "${GREEN}Seeding invoices into $(POSTGRES_DB)...${RESET}"
	@if command -v psql >/dev/null 2>&1; then \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-invoices.sql; \
	else \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-invoices.sql; \
	fi
	@echo "${GREEN}Invoices seed complete${RESET}"

seed-coupons: ## Load demo coupons for admin discounts page
	@echo "${GREEN}Seeding coupons into $(POSTGRES_DB)...${RESET}"
	@if command -v psql >/dev/null 2>&1; then \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-coupons.sql; \
	else \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-coupons.sql; \
	fi
	@echo "${GREEN}Coupons seed complete${RESET}"

seed-nav-menus-i18n: ## Load fa/es translations for storefront nav menus
	@echo "${GREEN}Seeding nav menu i18n into $(POSTGRES_DB)...${RESET}"
	@if command -v psql >/dev/null 2>&1; then \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-nav-menus-i18n.sql; \
	else \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-nav-menus-i18n.sql; \
	fi
	@echo "${GREEN}Nav menu i18n seed complete${RESET}"

seed-catalog-i18n: ## Load fa/es translations for products and categories
	@echo "${GREEN}Seeding catalog i18n into $(POSTGRES_DB)...${RESET}"
	@if command -v psql >/dev/null 2>&1; then \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-catalog-i18n.sql; \
	else \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-catalog-i18n.sql; \
	fi
	@echo "${GREEN}Catalog i18n seed complete${RESET}"

seed-orders-returns: ## Load demo orders + returns only (requires catalog seed)
	@echo "${GREEN}Seeding orders & returns into $(POSTGRES_DB)...${RESET}"
	@if command -v psql >/dev/null 2>&1; then \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-orders-returns.sql; \
	else \
		docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < scripts/seed-orders-returns.sql; \
	fi
	@echo "${GREEN}Orders & returns seed complete${RESET}"

clean: ## Clean build artifacts
	@echo "${YELLOW}Cleaning build artifacts...${RESET}"
	rm -rf $(GOBIN)
	@echo "${GREEN}Clean complete${RESET}"

## Database Migrations (goose)
migrate-create: ## Create a new migration file (usage: make migrate-create name=create_users_table)
	@if [ -z "$(name)" ]; then \
		echo "${YELLOW}Error: Migration name required${RESET}"; \
		echo "${GREEN}Usage: make migrate-create name=your_migration_name${RESET}"; \
		exit 1; \
	fi
	@echo "${GREEN}Creating migration: $(name)${RESET}"
	@$(GOOSE_CMD) create $(name) sql

migrate-up: ## Apply all pending migrations
	@echo "${GREEN}Running migrations against $(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB) as $(POSTGRES_USER)...${RESET}"
	@$(GOOSE_CMD) up || { \
		if [ "$(USE_COMPOSE)" = "1" ]; then \
			echo "${YELLOW}Host goose failed — retrying via Docker compose migrate profile${RESET}"; \
			$(MAKE) migrate-up-docker; \
		else \
			echo "${YELLOW}Migration failed. Check: make db-info${RESET}"; \
			exit 1; \
		fi; \
	}
	@echo "${GREEN}Migrations completed${RESET}"

migrate-up-docker: docker-wait-postgres ## Run goose inside compose network (USE_COMPOSE=1)
	@echo "${GREEN}Running migrations via compose migrate profile...${RESET}"
	COMPOSE_PROJECT_NAME=$(COMPOSE_PROJECT_NAME) $(COMPOSE) -f $(COMPOSE_FILE) --profile migrate run --rm migrate
	@echo "${GREEN}Migrations completed${RESET}"

migrate-down: ## Rollback the last migration
	@echo "${YELLOW}Rolling back last migration...${RESET}"
	@$(GOOSE_CMD) down
	@echo "${GREEN}Rollback completed${RESET}"

migrate-reset: ## Rollback all migrations (dangerous in production)
	@echo "${YELLOW}Resetting all migrations...${RESET}"
	@$(GOOSE_CMD) reset
	@echo "${GREEN}Reset completed${RESET}"

migrate-status: ## Check migration status
	@echo "${GREEN}Migration status:${RESET}"
	@$(GOOSE_CMD) status

migrate-version: ## Show current migration version
	@echo "${GREEN}Current migration version:${RESET}"
	@$(GOOSE_CMD) version

migrate-force: ## Force set migration version (usage: make migrate-force version=20250101120000)
	@if [ -z "$(version)" ]; then \
		echo "${YELLOW}Error: Version required${RESET}"; \
		echo "${GREEN}Usage: make migrate-force version=20250101120000${RESET}"; \
		exit 1; \
	fi
	@$(GOOSE_CMD) force $(version)

## Dependencies and Tools
deps: ## Download Go module dependencies
	@echo "${GREEN}Downloading dependencies...${RESET}"
	go mod download
	@echo "${GREEN}Dependencies downloaded${RESET}"

tidy: ## Tidy up go.mod and go.sum
	@echo "${GREEN}Tidying go modules...${RESET}"
	go mod tidy
	@echo "${GREEN}Complete${RESET}"

install-tools: ## Install development tools (goose, air)
	@echo "${GREEN}Installing goose...${RESET}"
	go install github.com/pressly/goose/v3/cmd/goose@v3.22.1
	@echo "${GREEN}Installing air (hot reload)...${RESET}"
	go install github.com/cosmtrek/air@latest
	@echo "${GREEN}Tools installed${RESET}"

## Testing
test: ## Run all tests
	@echo "${GREEN}Running tests...${RESET}"
	go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "${GREEN}Running tests with coverage...${RESET}"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}Coverage report generated: coverage.html${RESET}"

lint: ## Run go vet on all packages
	@echo "${GREEN}Running go vet...${RESET}"
	go vet ./...
	@echo "${GREEN}Lint complete${RESET}"

## Utility
watch: ## Run with hot reload (requires air)
	@command -v air >/dev/null 2>&1 || { \
		echo "${YELLOW}air not found, installing...${RESET}"; \
		go install github.com/cosmtrek/air@latest; \
	}
	air

## Full Setup
setup: install-tools deps migrate-up ## Install tools, dependencies, and run migrations
	@echo "${GREEN}Setup complete. You can now run 'make run'${RESET}"

reset-db: migrate-reset migrate-up ## Full reset: rollback all migrations and reapply
	@echo "${GREEN}Database reset complete${RESET}"

swagger: ## Generate Swagger documentation
	@echo "${GREEN}Generating Swagger docs...${RESET}"
	swag init --parseDependency --parseInternal --parseDepth 3 --overridesFile .swaggo -g cmd/api/main.go
	@echo "${GREEN}Swagger docs generated${RESET}"
