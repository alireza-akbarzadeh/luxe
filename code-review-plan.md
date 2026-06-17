# Code Review Plan

> **Note for AI / contributors:** Several findings below are outdated (graceful shutdown, health checks, Sentry, OTEL, integration tests, Swagger, and more already exist).
> **Use `.cursorrules`, `AGENTS.md`, and `documentation/architecture.md` for current standards.**

## Overview
This document captures the main code quality findings for the `go-shopping` project and recommends targeted improvements to take the project to the next level.

## Key Findings

### 1. Error handling is inconsistent
- Controllers duplicate error mapping for `AppError` versus generic errors.
- Some handlers use `ErrorResponse` directly while others use `InternalServerErrorResponse`.
- Service layer returns `utils.AppError`, but several controller branches still treat errors as plain strings.

### 2. Validation and request binding are repeated
- Many controllers perform `ShouldBindJSON` and validator checks in the same way.
- This duplication increases maintenance cost and makes request handling brittle.
- A reusable request-binding and validation layer would reduce boilerplate.

### 3. Business logic leaks into handlers
- Controllers still parse query strings, create filter objects, and enforce pagination defaults.
- Services should accept domain-specific filter structs or parameter objects from a helper layer.

### 4. Hard-coded domain values
- Cart/order statuses like `active`, `converted`, and `abandoned` appear as raw strings.
- Prefer typed constants or enums for status values to prevent spelling drift.

### 5. Startup and configuration could be more robust
- `cmd/api/main.go` uses `panic(...)` for boot failures.
- No signal handling / graceful shutdown workflow exists in the core startup flow.
- `config.Load()` validates only `JWT_SECRET`; other env fields should also be validated.

### 6. Missing tests and documentation
- The repository has no Go test files.
- There is no dedicated developer documentation for architecture, coding standards, or contribution guidance.

## Recommended Fix Categories

### A. Refactor and clean up
- Create centralized request binding + validation middleware.
- Add a centralized error mapper that converts `AppError` into HTTP responses.
- Replace string status literals with constants in `models`, `services`, and `controllers`.
- Extract pagination/filter parsing helpers for reuse.

### B. Improve architecture
- Introduce service interfaces across modules and inject them consistently via container wiring.
- Separate route registration, request validation, and business logic more clearly.
- Consider a lightweight hexagonal style for the next refactor: controllers → services → repositories.

### C. Strengthen stability and security
- Add graceful shutdown using signal handling and Gin server shutdown.
- Harden configuration validation and require production safe defaults.
- Remove any dev secret defaults from production documentation.

### D. Add test coverage
- Start with unit tests for service layer logic (`cart_service`, `orders_service`, `auth_service`).
- Add integration tests for API routes using Gin test server or `httptest`.
- Add regression tests for validation and error response handling.

## Candidate High-Impact Tasks

1. Build a `controllers/binder.go` helper that binds and validates request structs.
2. Add `pkg/http/response` or `utils/response` helper for unified API response formatting.
3. Create status constants in `constants` or `models`.
4. Add a `startup` package to manage config load, logger init, DB connect, worker pool, cron, and graceful shutdown.
5. Add a `docs/roadmap.md` and `docs/junior-guidance.md` for long-term maintenance.

## Short-term wins

- Add `go test` coverage for one service and one controller.
- Standardize validation error payloads across all controllers.
- Replace raw `panic()` in `main.go` with `logger.Fatal()` and proper shutdown.
- Document API error response contract in markdown.

## Estimated impact

High impact areas:
- `controllers/*` duplication
- `services/*` business rules with hard-coded strings
- startup/config / production readiness
- missing tests and docs

Low effort, high value:
- create common error handling middleware
- add request validator helper
- document current architecture and next steps
