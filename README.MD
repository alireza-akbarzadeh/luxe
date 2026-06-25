# Luxe API — Shopping Platform Backend

Production Go e-commerce API: **Gin**, **GORM**, **PostgreSQL**, **JWT**, optional **Redis/Asynq**, **Stripe**, **Sentry**, **OpenTelemetry**.

## Quick start

```bash
cp .env.example .env          # set JWT_SECRET, DB_*, etc.
make dev-setup                # Docker Postgres + Redis, migrations
make run                      # http://localhost:8080
```

Swagger UI: `http://localhost:8080/swagger/index.html` · OpenAPI: `http://localhost:8080/openapi`

More setup (Docker, migrations, AI, tracing): see sections below.

---

## Architecture (read this first)

The codebase uses **layered hexagonal architecture**. New work must follow these layers — do not put business logic or GORM back into HTTP handlers or grow monolithic services.

```
HTTP request
  → interfaces/http/handlers     (bind DTO, validate, utils.Response)
  → services/*_service.go        (thin facade — implements handler interfaces)
  → application/<context>/       (commands, queries, use cases)
  → domain/<context>/            (rules, ports, invariants)
  → infrastructure/postgres/     (GORM repositories)
  → PostgreSQL
```

Background work: `infrastructure/asynq` (Redis or in-memory) · cron: `internal/jobs/`

### Boot wiring

```
cmd/api/main.go
  → bootstrap.NewServices()      internal/application/bootstrap/wire.go
  → handlers.NewContainer()      internal/interfaces/http/handlers/container.go
  → routes.NewRouter().Setup()   internal/interfaces/http/routes/
```

Register every new domain in **`bootstrap/wire.go`** and **`handlers/container.go`**.

### Folder map

| Path | Purpose |
|------|---------|
| `cmd/api/` | Application entry, Swagger metadata |
| `internal/interfaces/http/handlers/` | Gin handlers (HTTP only) |
| `internal/interfaces/http/dto/` | Request/response DTOs |
| `internal/interfaces/http/routes/` | Route groups |
| `internal/interfaces/http/middleware/` | Auth, CORS, logging, tracing |
| `internal/services/` | Thin facades + `registry.go` (job handlers, workflow sync) |
| `internal/application/` | Use cases per bounded context (`cart`, `order`, `brand`, …) |
| `internal/application/bootstrap/` | Composition root (`NewServices`) |
| `internal/domain/` | Domain rules and repository **ports** (no GORM/Gin) |
| `internal/infrastructure/postgres/` | GORM repository implementations |
| `internal/infrastructure/asynq/` | Job queue |
| `internal/infrastructure/integrations/` | Stripe, R2, AI, PDF |
| `internal/infrastructure/workflow/` | Workflow engine |
| `internal/models/` | GORM entity structs (DB shape) |
| `internal/migrations/` | Goose SQL migrations |
| `internal/constants/` | Statuses, roles, route paths — no magic strings |
| `internal/shared/utils/` | Response helpers, logger, JWT |
| `internal/shared/observability/` | Sentry, OTEL, metrics |
| `internal/shared/health/` | Liveness/readiness |
| `tests/integration/` | Postgres integration tests |

Per-layer READMEs: `internal/application/README.md`, `internal/domain/README.md`, `internal/infrastructure/postgres/README.md`.

Full design doc: [`documentation/architecture.md`](documentation/architecture.md).

---

## How to add a feature

### New database table + REST resource

Follow **`.cursor/skills/new-api-entity/SKILL.md`** (or invoke `/new-api-entity` in Cursor):

1. `make migrate-create name=create_{entity}_table` → `make migrate-up` (schema only)
2. `internal/models/{entity}_model.go`
3. `internal/interfaces/http/dto/{entity}_dto.go`
4. `internal/infrastructure/postgres/{entity}_repository.go`
5. `internal/application/{entity}/` — `commands.go`, `queries.go`, mappers
6. `internal/services/{entity}_service.go` — thin facade delegating to application
7. Register in `application/bootstrap/wire.go` + `handlers/container.go`
8. `handlers/{entity}_handler.go` + `routes/{entity}_routes.go`
9. Swagger comments on handler → `make swagger` → **restart API**
10. **luxe-front:** `pnpm api:gen` → `pnpm check` (mandatory if OpenAPI changed)

### New route on an existing domain

Follow **`.cursor/skills/add-api-endpoint/SKILL.md`** (`/add-api-endpoint`):

- DTO in `interfaces/http/dto/` if shapes are new
- Application method (or extend existing commands/queries)
- Handler method + route + Swagger
- Tests + Swagger regen + frontend `api:gen` if contract changed

### Layer rules (non‑negotiable)

| Layer | Allowed | Forbidden |
|-------|---------|-----------|
| **Handlers** | Bind, validate, call service facade, `utils.Response` | Business logic, `*gorm.DB`, transactions |
| **Service facade** | Delegate to application; workflow sync; checkout orchestration | New GORM queries (add a repo instead) |
| **Application** | Orchestrate use cases, call repos + domain | Gin, HTTP, raw SQL in handlers |
| **Domain** | Rules, ports, pure types | GORM, Gin, framework imports |
| **Postgres repos** | GORM queries, mapping to domain/DTO | HTTP, business rules |

- **Statuses/roles** → `internal/constants`
- **Multi-table writes** → `db.Transaction` in application or checkout (not ad-hoc goroutines for orders/shipments)
- **Async order/shipment work** → `infrastructure/asynq.JobQueue`
- **Never edit** generated `docs/` — run `make swagger`
- **Migrations:** schema only — no demo product seed data in new Goose files

Checkout is the one intentional exception: `checkout_service.go` still coordinates a multi-service DB transaction.

---

## AI agents & Cursor

Before any task, read:

1. **`AGENTS.md`** — entry points and commands  
2. **`.cursorrules`** / **`.cursor/rules/luxe-go.mdc`** — hard conventions  
3. **`documentation/architecture.md`** — request flow and modules  
4. **Skills:** `.cursor/skills/new-api-entity`, `add-api-endpoint`

**Do not** recreate `internal/controllers/`, `internal/routes/`, or put GORM in handlers. **Do not** start another repo-wide refactor — extend the existing layers.

---

## Commands

| Command | Purpose |
|---------|---------|
| `make dev-setup` | Docker Postgres + Redis, migrate |
| `make run` | Start API |
| `make watch` | Hot reload (Air) |
| `make migrate-create name=...` | New Goose migration |
| `make migrate-up` / `migrate-down` | Apply / rollback |
| `make swagger` | Regenerate OpenAPI (`docs/` — gitignored) |
| `make test` | Unit tests |
| `make seed-dev` | Local demo data (not production) |

Integration tests:

```bash
DATABASE_URL=postgresql://postgres:postgres@localhost:5433/shopping_platform?sslmode=disable \
JWT_SECRET=ci_test_secret \
go test ./tests/integration/...
```

---

## Environment

See `.env.example`. Required: `JWT_SECRET`, database vars. Optional: `REDIS_URL`, `SENTRY_DSN`, `OTEL_*`, Stripe, R2, AI (`AI_ENABLED`, `AI_BASE_URL`, …).

---

## Frontend contract

OpenAPI is cached at API startup. After any Swagger/DTO/route change:

1. `make swagger`
2. Restart API
3. In **luxe-front**: `pnpm api:gen` → `pnpm check`

---

## Tech stack

Go 1.23+ · Gin · GORM · PostgreSQL 16+ · Goose · JWT · Validator · Viper · Logrus · Docker

---

## License

[MIT](LICENSE)
