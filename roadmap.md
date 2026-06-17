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

## Completed this session
- Controller deduplication: `paginationParams` + `parseUintParam` helpers in `controllers/helpers.go`; applied across `wallet`, `order`, `review`, `account`, `brand`, `category`, `address`, `coupon` controllers.
- Admin API: `GET /admin/stats`, `GET /admin/users`, `PATCH /admin/users/:id/role`, `PATCH /admin/users/:id/active` — all behind `RequireAdmin` middleware.
- RBAC integration tests: 401 (no token), 403 (regular user), 200 (admin) for `/admin/stats`.
- Admin service unit tests: `GetStats`, `ListUsers`, `UpdateUserRole` (valid + invalid), `ToggleUserActive`.
- Swagger regenerated.

## Completed (long-term)
- [x] Bulk admin ops: `POST /admin/orders/bulk-status` (up to 500 IDs) + `GET /admin/orders/export` (CSV, 10k rows, date/status filters).
- [x] Async email: `TypeSendEmail` task added to `JobQueue` interface; password-reset and verification emails now enqueued (Asynq when Redis available, memory queue fallback with inline goroutine safety net).

## Completed (long-term) — continued
- [x] Rate limiting: `StrictRateLimit` + `StandardRateLimit` named presets; TTL eviction prevents memory leak; applied to auth + import routes.
- [x] Webhook event log: `webhook_events` table, idempotency check on every Stripe delivery, per-event status lifecycle (`received → processed | failed`), `GET /admin/webhooks` list endpoint.
- [x] Import service unit tests: 9 table-driven tests covering category/product import (happy path, empty-name skip, service error, invalid price, store-ID fallback) and template generation.

## Completed (long-term) — continued
- [x] User order cancellation: `POST /orders/:id/cancel` — validates ownership + cancellable status (`pending`/`paid`), restores stock per item inside a transaction, refunds wallet-paid orders, cancels any pending shipment.
- [x] Wallet payment deduction: fixed `ProcessOrder` to call `walletService.DeductForOrder` when `payment.Method == "wallet"` instead of incorrectly routing through mock card gateway.
- [x] Order status emails: `UpdateOrderStatus` now enqueues async emails (via job queue) for `shipped`, `delivered`, and `cancelled` events using the user's preloaded email address.
- [x] `webhook_event_service` cleanup: replaced hand-rolled `containsStr`/`stringContains` helpers with `strings.Contains`.

## Next focus (long-term)
- Domain modularization (group by domain, not layer).

---

## Phase 6: Workflow state machine

### 6A — Engine & data (done)
- [x] DB-driven engine: workflows, states, transitions, audit logs, guards, hooks.
- [x] Seed definitions for order, product, shipment, return, user lifecycles.
- [x] Workflow CRUD + generic API (`/workflows/*`, `/admin/workflows/*`).
- [x] Makefile/DB setup for existing `docker-psql_bp-1` + `shopping_platform` database.

### 6B — Service integration (done)
- [x] Order, product, shipment, checkout, return, user lifecycle sync.
- [x] Event-driven `Transition` with `SetState` fallback + hooks (cancel, paid, shipped, publish, etc.).
- [x] Return domain: `ReturnService` + customer/admin routes.

### 6C — Admin transition APIs (done)
- [x] Product: `GET/POST /products/:id/available-transitions|transition`
- [x] Return: `GET/POST /admin/returns`, `POST /admin/returns/:id/transition`
- [x] Order: `GET/POST /orders/:id/available-transitions|transition`
- [x] Shipment: `GET/POST /shipments/:id/available-transitions|transition`
- [ ] Deprecate legacy `PUT /orders/:id/status` and `PUT /shipments/:id/status` once admin UI uses transitions

### 6D — Tests & docs (next)
- [ ] Integration tests: order cancel, return refund, product publish (real Postgres, skip if no `DATABASE_URL`)
- [x] Unit tests: `mirrorStatus`, role checks, nil-engine sync helpers
- [ ] Document workflow in `documentation/architecture.md`
- [ ] Regenerate Swagger (`make swagger`)

### 6E — Frontend (luxe-front)
- [ ] Shared hook: `useWorkflow(key)` → definition + state colors
- [ ] Admin product table: status badge from workflow state
- [ ] Admin product detail: action buttons from `available-transitions`
- [ ] Admin order detail: same pattern
- [ ] Optional: workflow history timeline component (`GET /workflows/:key/:id/history`)

### Suggested order of work
1. **Integration tests (6D)** — lock in cancel/refund/publish/deliver flows before frontend work.
2. **Workflow docs + Swagger (6D)** — `documentation/architecture.md` + `make swagger`.
3. **Frontend badges + actions (6E)** — highest user-visible value.
4. **Retire legacy status PUT** — once admin UI uses transitions everywhere.
