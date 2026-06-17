# Project Roadmap

> **Note for AI / contributors:** This file is a historical planning doc. Many items below are already done or superseded.
> **Use `.cursorrules`, `AGENTS.md`, and `documentation/architecture.md` for current conventions.**
> Verify the codebase before treating unchecked boxes as still required (e.g. repository layer is transitional, not a goal).

## Goal
Bring `go-shopping` from a working backend project to a production-ready, maintainable e-commerce platform.

## Roadmap Phases

### Phase 1: Stabilize and standardize
- [ ] Add centralized request validation and binding helpers.
- [ ] Create a shared error handling layer for `AppError` → HTTP response mapping.
- [ ] Replace raw string statuses (`active`, `converted`, `abandoned`, `pending`) with typed constants.
- [ ] Refactor controllers to reduce duplication and improve readability.
- [ ] Remove any developer-only config defaults from production paths.

### Phase 2: Add test coverage and documentation
- [ ] Add unit tests for key services: `cart_service`, `orders_service`, `auth_service`.
- [ ] Add route tests for at least cart and order endpoints.
- [ ] Add developer docs: architecture overview, component responsibilities, API conventions.
- [ ] Add `Makefile` targets for `test`, `test-coverage`, and `lint`.

### Phase 3: Harden platform reliability
- [ ] Implement graceful shutdown for Gin + worker pool + DB connection.
- [ ] Add health endpoints and readiness probes.
- [ ] Add DB connection retry logic and proper connection pooling settings.
- [ ] Add validation for required environment variables, and fail fast if invalid.

### Phase 4: Improve architecture and extensibility
- [ ] Introduce a repository/data access layer to isolate GORM from services.
- [ ] Implement request/response DTOs for public API contracts.
- [ ] Add OpenAPI/Swagger docs generation from source annotations.
- [ ] Add role-based permissions middleware and use it consistently.

### Phase 5: Production feature polish
- [ ] Add full order lifecycle: payment status, shipment tracking, cancellation.
- [ ] Add inventory reservation and stock locking to prevent overselling.
- [ ] Add customer account features: address book, order history filters, profile update.
- [ ] Add admin dashboards or a lightweight admin API.

## Quick wins
- Standardize JSON response format across all endpoints.
- Add a `docs/architecture.md` or `docs/system-design.md` file.
- Add API version prefix consistently in routes (e.g. `/api/v1`).
- Add one `README` section for local dev and Docker workflows.

## Long-term vision
- Modularize by domain: `cart`, `orders`, `products`, `users`, `payments`, `shipments`.
- Add event-driven or async processing for order fulfillment and inventory updates.
- Add observability: structured logging, request tracing, metrics, error reporting.
- Add CI/CD pipeline with linting, tests, migration checks, and deployment previews.

## Milestones
1. `M1` — Basic stability: unified responses, no duplicate validation, `go test` coverage.
2. `M2` — Production readiness: graceful shutdown, config validation, health probes.
3. `M3` — Feature readiness: checkout resiliency, inventory safety, admin pages.
4. `M4` — Observability and deployment: metrics, logging, CI/CD.


testing strategy
shopping-platform/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── controllers/
│   │   ├── auth.go
│   │   └── auth_test.go     // ✅ test for auth.go
│   ├── services/
│   │   ├── auth_service.go
│   │   └── auth_service_test.go // ✅ test for auth_service.go
│   ├── models/
│   │   ├── user.go
│   │   └── user_test.go     // ✅ test for user.go
│   ├── dto/
│   │   ├── product.go
│   │   └── product_test.go  // ✅ test for product.go
│   └── middleware/
│       ├── auth.go
│       └── auth_test.go     // ✅ test for auth.go
├── tests/                    // 🧪 Reserved for integration & E2E tests
│   ├── integration/
│   └── e2e/
├── go.mod
└── go.sum