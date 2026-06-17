# AGENTS.md — Luxe API

Guide for AI agents and automated tools working in this repository.

## Project

Production Go e-commerce API: **Gin**, **GORM**, **PostgreSQL**, **JWT**, optional **Redis/Asynq**, **Stripe**, **Sentry**, **OpenTelemetry**.

## Authoritative rules

1. **`.cursorrules`** — full conventions (layers, migrations, swagger, tests).
2. **`.cursor/rules/luxe-go.mdc`** — short always-on summary.
3. **`documentation/architecture.md`** — system design and wiring.

Do not follow outdated items in `roadmap.md` or `code-review-plan.md` without verifying the codebase.

## Architecture (one line)

```
HTTP → Controller → Service (*gorm.DB + business logic) → PostgreSQL
```

- Controllers: bind DTO, validate, `utils.Response`, pass `c.Request.Context()`.
- Services: rules, transactions, orchestration, `utils.Err*`.
- Do **not** add `internal/repositories/` — data access lives in services via `*gorm.DB`.

## New feature checklist

1. `make migrate-create name=...` (schema only) → `make migrate-up`
2. `internal/models/` → `internal/dto/` → `internal/services/` (+ `services.go`)
3. `internal/controllers/` (+ `container.go`) → `internal/routes/`
4. Swagger comments on handlers → `make swagger`
5. Tests: unit (`internal/services/*_test.go`) or `tests/integration/`

## Commands

| Command | Purpose |
|---------|---------|
| `make dev-setup` | Docker Postgres + Redis, wait, migrate (Goose CLI) |
| `make migrate-up-docker` | Goose via Docker network (Windows Docker fallback) |
| `make run` | Start API |
| `make migrate-create name=<name>` | New migration |
| `make migrate-up` | Apply migrations |
| `make swagger` | Regenerate `docs/` (gitignored) |
| `make seed-dev` | Load dev demo data (local only) |
| `make test` | Unit tests |

Integration tests:

```bash
DATABASE_URL=postgresql://postgres:postgres@localhost:5433/shopping_platform?sslmode=disable \
JWT_SECRET=ci_test_secret \
go test ./tests/integration/...
```

## Never edit

- `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml` (generated)
- `.env` (secrets)

## Key entry points

| File | Role |
|------|------|
| `cmd/api/main.go` | Boot, observability, cron, job queue |
| `internal/services/services.go` | `NewServices` DI |
| `internal/controllers/container.go` | `NewContainer` |
| `internal/routes/setup.go` | Route groups, middleware |
| `internal/config/` | Env loading, production validation |

## Env reference

See `.env.example` for `DATABASE_URL`, `JWT_SECRET`, `REDIS_URL`, `SENTRY_DSN`, `OTEL_*`, Stripe, R2.
