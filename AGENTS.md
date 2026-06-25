# AGENTS.md — Luxe API

Guide for AI agents and automated tools working in this repository.

## Project

Production Go e-commerce API: **Gin**, **GORM**, **PostgreSQL**, **JWT**, optional **Redis/Asynq**, **Stripe**, **Sentry**, **OpenTelemetry**.

## Authoritative rules

1. **`README.md`** — architecture overview and how to add features (start here).
2. **`.cursorrules`** — full conventions (layers, migrations, swagger, tests).
3. **`.cursor/rules/luxe-go.mdc`** — short always-on summary (Cursor Rules).
4. **`documentation/architecture.md`** — system design and wiring.
5. **`.cursor/skills/`** — on-demand workflow skills (see below); invoke with `/skill-name` or let Agent auto-load.
6. **`documentation/FEATURES_PLAN.md`** — product roadmap: what exists, gaps, priorities, AI plan.

Do not follow outdated items in `roadmap.md` without verifying the codebase. **Do not** recreate deleted legacy paths (`internal/controllers/`, `internal/routes/`, `internal/dto/`, `internal/utils/`, `internal/middleware/`, `internal/tasks/`).

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
| `/find-skills` | Find/install agent skills — local skills first, then skills.sh |
| `/new-api-entity` | New table + full domain — migration → model → DTO → service → controller → routes → Swagger |
| `/add-api-endpoint` | New handler on an **existing** service (bulk action, extra route) — no new entity |

Details: `.cursor/skills/README.md`. Frontend follow-up after Swagger: **`luxe-front`** → restart API first → `/api-gen` → `pnpm api:gen`.

**Swagger → frontend rule:** Any OpenAPI contract change (new/renamed DTO, field, route, response shape) requires **`make swagger` → restart API → `pnpm api:gen` in luxe-front**. `make swagger` alone does not update the frontend.

**Boundary:** new database table from scratch → `/new-api-entity`. Single route on existing `*_service.go` → `/add-api-endpoint`.

## Architecture (one line)

```
Handler → services facade → application → domain → infrastructure/postgres → PostgreSQL
```

- Composition root: `internal/application/bootstrap/wire.go` (`bootstrap.NewServices`).
- Handlers: `internal/interfaces/http/handlers/` — no business logic, no GORM.
- **Extend existing layers** — do not start another repo-wide refactor or reintroduce `controllers/` / monolithic GORM services.

## New feature checklist

1. Decide scope: **new entity** → `/new-api-entity`; **extra route on existing service** → `/add-api-endpoint`.
2. `make migrate-create name=...` (schema only) → `make migrate-up` — if new table required.
3. `internal/models/` → `internal/interfaces/http/dto/` → `internal/application/<ctx>/` + `internal/infrastructure/postgres/` → thin `internal/services/*_service.go` facade (+ **`bootstrap/wire.go`**).
4. `internal/interfaces/http/handlers/` (+ **`container.go`**) → `internal/interfaces/http/routes/`.
5. Swagger comments on handlers → `make swagger` (never edit `docs/` by hand).
6. Tests: unit (`internal/services/*_test.go`) or `tests/integration/`.
7. Restart API; in **luxe-front**: `pnpm api:gen` + `pnpm check` (mandatory if DTOs, routes, or Swagger comments changed).
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
| `internal/application/bootstrap/wire.go` | `bootstrap.NewServices` DI |
| `internal/services/registry.go` | `services.Services` type, `JobHandlers` |
| `internal/interfaces/http/handlers/container.go` | `NewContainer` |
| `internal/interfaces/http/routes/setup.go` | Route groups, middleware |
| `internal/config/` | Env loading, production validation |
| `.cursor/skills/` | Agent Skills — `/new-api-entity`, `/add-api-endpoint` |
| `.cursor/skills/README.md` | Skill index, evals, triggering tests |

## Env reference

See `.env.example` for `DATABASE_URL`, `JWT_SECRET`, `REDIS_URL`, `SENTRY_DSN`, `OTEL_*`, Stripe, R2.
