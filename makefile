# Makefile for shopping-platform (PostgreSQL + Goose)

# Variables
BINARY_NAME=shopping-platform-api
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
MIGRATIONS_DIR=./internal/migrations

# Database DSN – defaults to environment variable, falls back to dev settings
# Override with: make migrate-up DATABASE_URL="postgres://user:pass@localhost:5432/db?sslmode=disable"
DATABASE_URL ?= postgresql://postgres:postgres@localhost:5433/shopping_platform?sslmode=disable

# Goose CLI (install: make install-tools or go install github.com/pressly/goose/v3/cmd/goose@latest)
GOOSE_CMD=goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)"
GOOSE_DOCKER_URL=postgresql://postgres:postgres@postgres:5432/shopping_platform?sslmode=disable

# Colors for output
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: help build run clean test migrate-create migrate-up migrate-up-docker migrate-down migrate-reset migrate-status migrate-force deps tidy install-tools docker-up docker-up-jaeger docker-wait-postgres stripe-listen dev-setup seed-dev

# Default target
help: ## Show this help message
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${YELLOW}%-20s${RESET} ${GREEN}%s${RESET}\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

## Development
build: ## Build the application
	@echo "${GREEN}Building application...${RESET}"
	go build -o $(GOBIN)/$(BINARY_NAME) ./cmd/api
	@echo "${GREEN}Build complete: $(GOBIN)/$(BINARY_NAME)${RESET}"

run: ## Run the application (uses .env or environment variables)
	@echo "${GREEN}Running application...${RESET}"
	go run ./cmd/api

docker-up: ## Start PostgreSQL and Redis (required for local dev)
	@echo "${GREEN}Starting PostgreSQL and Redis...${RESET}"
	docker compose up -d postgres redis
	@echo "${GREEN}Core services ready. For tracing UI: make docker-up-jaeger (optional)${RESET}"
	@echo "${YELLOW}Tip: wait for Postgres before migrate-up — use make dev-setup or make docker-wait-postgres${RESET}"

docker-wait-postgres: ## Wait until Docker Postgres accepts connections
	@echo "${GREEN}Waiting for Postgres...${RESET}"
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30; do \
		docker compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1 && { echo "${GREEN}Postgres ready${RESET}"; exit 0; }; \
		sleep 1; \
	done; \
	echo "${YELLOW}Postgres not ready after 30s — check docker compose ps${RESET}"; exit 1

docker-up-jaeger: ## Start Jaeger for OTLP tracing (optional; needs Docker Hub access)
	@echo "${GREEN}Starting Jaeger...${RESET}"
	docker compose up -d jaeger || (echo "${YELLOW}Jaeger failed to start (network/Docker Hub?). OTEL_ENABLED still works if a collector is reachable.${RESET}" && exit 1)

docker-up-all: docker-up docker-up-jaeger ## Start Postgres, Redis, and Jaeger

stripe-listen: ## Forward Stripe webhooks (orders + wallet deposits) to local API
	@echo "${GREEN}Forwarding Stripe webhooks to localhost:8080/api/v1/webhooks/stripe${RESET}"
	stripe listen --forward-to localhost:8080/api/v1/webhooks/stripe --events checkout.session.completed,checkout.session.expired

dev-setup: docker-up docker-wait-postgres migrate-up ## Start Postgres, Redis, and run migrations

seed-dev: ## Load dev demo data (local/staging only; uses psql or docker exec)
	@echo "${GREEN}Seeding dev demo data...${RESET}"
	@if command -v psql >/dev/null 2>&1; then \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed-dev.sql; \
	else \
		echo "${YELLOW}psql not found - seeding via docker (shopping_platform_postgres)${RESET}"; \
		docker exec -i shopping_platform_postgres psql -U postgres -d shopping_platform -v ON_ERROR_STOP=1 < scripts/seed-dev.sql; \
	fi
	@echo "${GREEN}Dev seed complete${RESET}"

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
	$(GOOSE_CMD) create $(name) sql

migrate-up: ## Apply all pending migrations
	@echo "${GREEN}Running migrations...${RESET}"
	@$(GOOSE_CMD) up || { \
		echo "${YELLOW}Host goose failed — retrying via Docker network (common on Windows Docker)${RESET}"; \
		$(MAKE) migrate-up-docker; \
	}
	@echo "${GREEN}Migrations completed${RESET}"

migrate-up-docker: docker-wait-postgres ## Run goose inside Docker network (when host port 5433 fails on Windows)
	@echo "${GREEN}Running migrations via Docker (external goose)...${RESET}"
	docker compose --profile migrate run --rm migrate
	@echo "${GREEN}Migrations completed${RESET}"

migrate-down: ## Rollback the last migration
	@echo "${YELLOW}Rolling back last migration...${RESET}"
	$(GOOSE_CMD) down
	@echo "${GREEN}Rollback completed${RESET}"

migrate-reset: ## Rollback all migrations (dangerous in production)
	@echo "${YELLOW}Resetting all migrations...${RESET}"
	$(GOOSE_CMD) reset
	@echo "${GREEN}Reset completed${RESET}"

migrate-status: ## Check migration status
	@echo "${GREEN}Migration status:${RESET}"
	$(GOOSE_CMD) status

migrate-version: ## Show current migration version
	@echo "${GREEN}Current migration version:${RESET}"
	$(GOOSE_CMD) version

migrate-force: ## Force set migration version (usage: make migrate-force version=20250101120000)
	@if [ -z "$(version)" ]; then \
		echo "${YELLOW}Error: Version required${RESET}"; \
		echo "${GREEN}Usage: make migrate-force version=20250101120000${RESET}"; \
		exit 1; \
	fi
	$(GOOSE_CMD) force $(version)

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