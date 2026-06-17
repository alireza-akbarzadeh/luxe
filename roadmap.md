# Project Roadmap

> **Note for AI / contributors:** This file tracks phase progress. For current conventions use `.cursorrules`, `AGENTS.md`, and `documentation/architecture.md`.
> Verify the codebase before treating unchecked boxes as still required (repository layer was removed — services + GORM is the target).

## Goal
Bring Luxe from a working backend to a production-ready, maintainable e-commerce platform.

## Roadmap Phases

### Phase 1: Stabilize and standardize
- [x] Add centralized request validation and binding helpers (`utils.BindAndValidate`).
- [x] Create a shared error handling layer for `AppError` → HTTP response mapping (`HandleServiceError`).
- [x] Replace raw string statuses with typed constants (cart, order, payment, wallet, product, store).
- [x] Refactor controllers to reduce duplication (`paginationParams`, `parseUintParam` in `controllers/helpers.go`).
- [x] Remove developer-only JWT default from production paths (`JWT_SECRET` default only when `APP_ENV=local`).

### Phase 2: Add test coverage and documentation
- [x] Add unit tests for key services: `auth_service`, `cart_service`, `orders_service`, `wallet_service`.
- [x] Add route/controller tests for auth endpoints.
- [x] Add developer docs: `documentation/architecture.md`, `AGENTS.md`.
- [x] Add `Makefile` targets for `test`, `test-coverage`, and `lint` (`go vet`).

### Phase 3: Harden platform reliability
- [x] Implement graceful shutdown for Gin + worker pool + DB connection (`cmd/api/server.go`).
- [x] Add health endpoints and readiness probes (`/api/v1/health/*`).
- [x] Add DB connection retry logic on startup (`database.ConnectWithRetry`).
- [x] Add validation for required environment variables, fail fast in production (`config.Validate`).

### Phase 4: Improve architecture and extensibility
- [x] ~~Introduce repository layer~~ — **not planned**; services use `*gorm.DB` directly.
- [x] Implement request/response DTOs for public API contracts.
- [x] Add OpenAPI/Swagger docs generation from source annotations.
- [x] Add role-based permissions middleware and use it consistently across admin routes (`RequireAdmin`, nav menu write protection).

### Phase 5: Production feature polish
- [x] Order lifecycle: payment (Stripe/mock/wallet), shipment tracking, cancellation paths.
- [x] Inventory checks at checkout; stock decrement on order processing.
- [x] Customer account: address book, order history filters, profile update.
- [x] Admin dashboards or expanded lightweight admin API (`GET /api/v1/admin/stats`, `AdminService`).

## Quick wins
- [x] Standardize JSON response format (`utils.Response`).
- [x] Architecture doc (`documentation/architecture.md`).
- [x] API version prefix `/api/v1`.
- [x] README local dev section.

## Long-term vision
- Modularize by domain (cart, orders, products, users, payments, shipments).
- Event-driven / async processing (Asynq job queue when `REDIS_URL` set).
- [x] Observability: structured logging, OTEL tracing, Prometheus `/metrics`, Sentry.
- [x] CI/CD pipeline with Postgres, Redis, migrations, tests.

## Milestones
1. `M1` — Basic stability: unified responses, validation helpers, service tests. **Mostly done**
2. `M2` — Production readiness: graceful shutdown, config validation, health probes. **Mostly done**
3. `M3` — Feature readiness: checkout resiliency, inventory safety, admin APIs. **In progress**
4. `M4` — Observability and deployment: metrics, logging, CI/CD. **Mostly done**

## Next focus (suggested order)
1. Phase 5 — Admin API expansion (reports, bulk ops).
2. Phase 1 — Controller deduplication (pagination helpers).
3. Integration tests for admin RBAC (403 for non-admin).
