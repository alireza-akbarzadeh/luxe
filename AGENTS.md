# AGENTS.md — Luxe API

Guide for AI agents and automated tools working in this repository.

## Project

Production Go e-commerce API: **Gin**, **GORM**, **PostgreSQL**, **JWT**, optional **Redis/Asynq**, **Stripe**, **Sentry**, **OpenTelemetry**.

## Authoritative rules

1. **`.cursorrules`** — full conventions (layers, migrations, swagger, tests).
2. **`.cursor/rules/luxe-go.mdc`** — short always-on summary (Cursor Rules).
3. **`documentation/architecture.md`** — system design and wiring.
4. **`.cursor/skills/`** — on-demand workflow skills (see below); invoke with `/skill-name` or let Agent auto-load.
5. **`documentation/FEATURES_PLAN.md`** — product roadmap: what exists, gaps, priorities, AI plan.

Do not follow outdated items in `roadmap.md` or `code-review-plan.md` without verifying the codebase.

## Cursor AI context (`.cursor/`)

Two layers — hard rules stay here and in `.cursorrules`; step-by-step workflows live in skills.

| Layer | Path | Role |
|-------|------|------|
| **Always-on rules** | `.cursorrules`, `.cursor/rules/luxe-go.mdc` | Hard conventions every session |
| **Agent skills** | `.cursor/skills/<name>/SKILL.md` | Task workflows loaded when relevant |
| **Skill evals** | `.cursor/skills/<name>/evals/`, `eval-queries.json` | Test description triggering and output quality |

### Project skills (luxe API)

| Skill | Use when |
|-------|----------|
| `/new-api-entity` | New table + full domain — migration → model → DTO → service → controller → routes → Swagger |
| `/add-api-endpoint` | New handler on an **existing** service (bulk action, extra route) — no new entity |

Details: `.cursor/skills/README.md`. Frontend follow-up after Swagger: **`luxe-front`** repo → `/api-gen` → `pnpm api:gen`.

**Boundary:** new database table from scratch → `/new-api-entity`. Single route on existing `*_service.go` → `/add-api-endpoint`.

## Architecture (one line)

```
HTTP → Controller → Service (*gorm.DB + business logic) → PostgreSQL
```

- Controllers: bind DTO, validate, `utils.Response`, pass `c.Request.Context()`.
- Services: rules, transactions, orchestration, `utils.Err*`.
- Do **not** add `internal/repositories/` — data access lives in services via `*gorm.DB`.

## New feature checklist

1. Decide scope: **new entity** → `/new-api-entity`; **extra route on existing service** → `/add-api-endpoint`.
2. `make migrate-create name=...` (schema only) → `make migrate-up` — if new table required.
3. `internal/models/` → `internal/dto/` → `internal/services/` (+ **`services.go`**).
4. `internal/controllers/` (+ **`container.go`**) → `internal/routes/`.
5. Swagger comments on handlers → `make swagger` (never edit `docs/` by hand).
6. Tests: unit (`internal/services/*_test.go`) or `tests/integration/`.
7. Restart API; tell frontend to run `pnpm api:gen` in **luxe-front**.
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
| `.cursor/skills/` | Agent Skills — `/new-api-entity`, `/add-api-endpoint` |
| `.cursor/skills/README.md` | Skill index, evals, triggering tests |

## Env reference

See `.env.example` for `DATABASE_URL`, `JWT_SECRET`, `REDIS_URL`, `SENTRY_DSN`, `OTEL_*`, Stripe, R2.
