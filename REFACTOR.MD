@codebase

## Mission
Refactor this Go + Gin + PostgreSQL ecommerce backend into a production-grade,
senior-level system using Clean Architecture, DDD, and Hexagonal Architecture.
Do NOT rewrite everything. Work phase by phase. Stop after each phase and wait
for my approval before continuing.

---

## Phase 0 — Audit (start here, do nothing else yet)

Scan the entire codebase and produce this exact report:

```
ARCHITECTURE AUDIT
==================
Score: X/10

Critical issues (fix now):
  [ ] ...

High priority (fix this sprint):
  [ ] ...

Medium priority (fix next sprint):
  [ ] ...

Low priority (tech debt backlog):
  [ ] ...

Domain boundaries identified:
  - User
  - Product
  - Order
  - Payment
  - ... (list all you find)

Files that violate Clean Architecture:
  - path/to/file.go → reason

Missing interfaces (need to add):
  - ...

Missing tests (coverage gaps):
  - ...

Recommended migration order:
  1. ...
  2. ...
```

After producing the audit, stop. Do not write any code yet.
Ask: "Does this audit look correct? Which phase should I start with?"

---

## Phase rules (apply to EVERY phase)

Before writing any code:
  - State: what problem this phase solves
  - State: migration risk (Low / Medium / High + reason)
  - State: what "done" looks like
  - State: rollback plan

While writing code:
  - One domain at a time
  - No file > 300 lines
  - Write tests alongside the code, not after
  - Never break existing API contracts without a migration path

After writing code:
  - Show the diff summary (which files changed, added, deleted)
  - Show how to run the tests: `go test ./...`
  - Update ARCHITECTURE.md if structure changed
  - Ask: "Phase complete. Ready for the next one?"

---

## Reference architecture
(See .cursorrules for full details)

cmd/ → entrypoints only
internal/domain/ → pure Go, no framework imports, all business logic
internal/application/ → orchestration, CQRS commands/queries
internal/infrastructure/ → implements domain interfaces, DB/Redis/Queue/etc
internal/interfaces/http/ → Gin handlers, middleware, DTOs
shared/ → config, errors, logger, telemetry

---

## Tech decisions (already made, don't change without asking)
- Router: Gin
- ORM: raw sqlx (no GORM)
- Migrations: goose
- Background jobs: Asynq
- Cache: Redis (cache-aside)
- Logging: slog or uber/zap (pick one, stay consistent)
- Metrics: Prometheus
- Tracing: OpenTelemetry
- Docs: swaggo (OpenAPI 3.0)
- Tests: testcontainers-go for integration
- CI: GitHub Actions (lint → test → build → security scan → docker)

---

## Start now with Phase 0.




# ── PHASE RUNNER TEMPLATE ──────────────────────────────────────
# Replace {PHASE} and {DOMAIN} with actual values before pasting.

## Phase {PHASE}: {TITLE}
@codebase

### What I want
{One sentence describing the goal}

### Acceptance criteria
- [ ] ...
- [ ] ...

### Files to create or modify
- internal/domain/{domain}/entity.go
- internal/domain/{domain}/repository.go
- internal/domain/{domain}/service.go
- internal/domain/{domain}/errors.go
- internal/domain/{domain}/events.go
- internal/domain/{domain}/service_test.go

### Constraints
- Do not touch files outside the {domain} boundary unless you ask first
- Handler stays untouched this phase — only domain + infra
- All new code must pass: go test ./internal/domain/{domain}/...
- Run golangci-lint before finishing

### Output format
1. List every file you'll change and why
2. Write the code
3. Write the tests
4. Show: go test -v ./internal/domain/{domain}/... output (simulated)
5. Ask for approval before moving on

# ── READY-TO-USE PHASE PROMPTS ─────────────────────────────────

## COPY THIS for Phase 1: Folder restructure
Restructure the codebase to match the target architecture in .cursorrules.
Move files only — do not refactor logic yet. Create empty placeholder files
with package declarations where needed. After moving, verify:
  go build ./...
shows zero errors. List every mv operation you perform.

## COPY THIS for Phase 2: Domain modeling (pick one domain per run)
Model the [Product] domain from scratch following DDD.
Create all 5 files (entity, repository interface, service, errors, events).
Use uuid.UUID for IDs. Use a Money value object for prices.
No database imports in the domain layer. No gin imports. Pure Go only.
Write table-driven unit tests for every service method.

## COPY THIS for Phase 3: Repository implementation
Implement the [Product]Repository interface using sqlx + Postgres.
File: internal/infrastructure/postgres/product_repository.go
Use parameterized queries only. Wrap every method with an OpenTelemetry span.
Write an integration test using testcontainers-go that spins up a real Postgres.

## COPY THIS for Phase 4: Handler refactor
Refactor the [product] Gin handlers to be < 60 lines each.
Move all business logic to the service layer first if it hasn't been.
Create DTOs in internal/interfaces/http/dto/product.go
Never return a domain model directly — map everything through the DTO.
Add swaggo annotations to every endpoint.

## COPY THIS for Phase 5: Observability
Add structured logging (slog), Prometheus metrics, and OpenTelemetry tracing
to the entire request lifecycle. Instrument:
  - Every HTTP handler (request_id, duration, status_code)
  - Every DB query (query name, duration, error)
  - Every Redis call
  - Every background job
Add /metrics endpoint for Prometheus scrape.
Add /health, /ready, /live endpoints with real dependency checks.