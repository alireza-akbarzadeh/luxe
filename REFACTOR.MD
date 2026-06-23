

# Goal

Transform an existing Go + Gin + PostgreSQL ecommerce backend into a production-grade, senior-level codebase that is:

* Easy to maintain
* Easy to scale
* Well tested
* Observable
* Secure
* Domain-driven
* Cloud-native
* CI/CD ready
* Developer friendly

---

# Phase 0: Architecture Assessment

## Task

Perform a complete architecture audit.

### Deliverables

1. Analyze current folder structure.
2. Identify:

   * MVC violations
   * Business logic inside controllers
   * Fat services
   * God structs
   * Circular dependencies
   * Duplicate code
   * Missing interfaces
   * Missing tests
   * Missing observability
3. Produce:

```text
CURRENT STATE REPORT

Architecture Score: X/10

Problems:
- ...
- ...
- ...

Recommended Changes:
- ...
```

---

# Phase 1: Define Target Architecture

Move from MVC toward:

```text
cmd/
internal/

├── domain/
│   ├── product/
│   ├── order/
│   ├── user/
│   └── payment/
│
├── application/
│   ├── commands/
│   ├── queries/
│   └── services/
│
├── infrastructure/
│   ├── postgres/
│   ├── redis/
│   ├── queue/
│   ├── payment/
│   └── storage/
│
├── interfaces/
│   ├── http/
│   ├── websocket/
│   └── grpc/
│
└── shared/
```

### Goal

Adopt:

* Clean Architecture
* Hexagonal Architecture
* Domain Driven Design (DDD)

without over-engineering.

---

# Phase 2: Domain Modeling

AI should identify core ecommerce domains.

Example:

```text
User
Product
Inventory
Cart
Order
Payment
Coupon
Review
Shipment
Notification
```

For each domain:

Create:

```go
entity.go
repository.go
service.go
errors.go
events.go
```

Example:

```go
type Product struct {
    ID          uuid.UUID
    Name        string
    Price       Money
    Stock       int
}
```

Keep business rules inside domain.

---

# Phase 3: Eliminate Fat Controllers

Controllers should only:

```text
Receive request
Validate request
Call service
Return response
```

Bad:

```go
func CreateOrder(c *gin.Context) {
    // 500 lines
}
```

Good:

```go
func (h *Handler) CreateOrder(c *gin.Context) {
    req := CreateOrderRequest{}

    if err := c.ShouldBindJSON(&req); err != nil {
        ...
    }

    order, err := h.orderService.Create(...)
}
```

Target:

```text
Controller < 100 lines
```

---

# Phase 4: Repository Pattern

Introduce interfaces.

Example:

```go
type ProductRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
    Create(ctx context.Context, p *Product) error
}
```

Implementation:

```go
postgres/
    product_repository.go
```

Benefits:

* Easier testing
* Easier migration
* Cleaner boundaries

---

# Phase 5: Database Layer Improvements

Review all migrations.

Audit:

### Indexes

```sql
CREATE INDEX idx_orders_user_id
ON orders(user_id);
```

### Constraints

```sql
UNIQUE
CHECK
FOREIGN KEY
```

### Soft Deletes

```sql
deleted_at
```

### Transaction Safety

Use:

```go
tx, err := db.BeginTx(...)
```

for:

* Checkout
* Inventory updates
* Payment operations

---

# Phase 6: Configuration Management

Current goal:

```text
12-Factor App
```

Create:

```go
internal/config
```

Use:

```env
DATABASE_URL=
REDIS_URL=
JWT_SECRET=
```

Validate on startup.

Example:

```go
type Config struct {
    DatabaseURL string
}
```

No magic values.

---

# Phase 7: Error Handling Strategy

Create standard errors.

Example:

```go
var (
    ErrProductNotFound
    ErrInsufficientStock
    ErrUnauthorized
)
```

Map errors globally.

```go
middleware/error_handler.go
```

Output:

```json
{
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "product not found"
  }
}
```

---

# Phase 8: Validation Layer

Use DTOs.

```text
dto/
```

Separate:

```go
Product
ProductRequest
ProductResponse
```

Never expose DB models directly.

---

# Phase 9: Authentication & Authorization

Review:

### JWT

* Rotation
* Refresh tokens

### RBAC

Roles:

```text
admin
customer
seller
support
```

Middleware:

```go
RequireRole(...)
```

---

# Phase 10: Observability

Many junior projects skip this.

Add:

### Structured Logging

Use:

```go
log/slog
```

or

```go
uber/zap
```

Example:

```go
logger.Info(
    "order created",
    "order_id", order.ID,
    "user_id", user.ID,
)
```

---

### Metrics

Add:

```text
Prometheus
```

Metrics:

```text
http_requests_total
order_created_total
payment_failed_total
```

---

### Tracing

Add:

```text
OpenTelemetry
```

Trace:

```text
HTTP
DB
Redis
External APIs
```

---

# Phase 11: Caching Strategy

Introduce Redis.

Use for:

```text
Product catalog
Categories
Sessions
Rate limiting
```

Implement cache-aside pattern.

---

# Phase 12: Background Jobs

Current jobs folder should be reviewed.

Use:

```text
Asynq
```

for:

```text
Emails
Notifications
Invoice generation
Inventory sync
```

Avoid goroutines in handlers.

Bad:

```go
go sendEmail(...)
```

Good:

```go
asynqClient.Enqueue(...)
```

---

# Phase 13: API Versioning

Introduce:

```text
/api/v1
/api/v2
```

Avoid breaking clients.

---

# Phase 14: API Documentation

Generate:

```text
OpenAPI 3.0
```

Use:

```text
swaggo
```

Every endpoint documented.

---

# Phase 15: Testing Strategy

Target:

```text
70-80% meaningful coverage
```

### Unit Tests

```go
service_test.go
```

Mock repositories.

### Integration Tests

Real:

```text
Postgres
Redis
```

via:

```text
Testcontainers
```

### E2E Tests

Checkout flow.

---

# Phase 16: Security Audit

Check:

### SQL Injection

Use:

```go
parameterized queries
```

### Secrets

Never commit:

```env
.env
```

### Security Headers

### Rate Limiting

Use:

```text
Redis
```

### CORS

Review carefully.

---

# Phase 17: Performance Audit

Benchmark:

```go
go test -bench
```

Review:

* N+1 queries
* Missing indexes
* Slow endpoints
* Memory leaks

Add:

```text
pprof
```

---

# Phase 18: CI/CD

GitHub Actions:

```yaml
lint
test
build
security scan
docker build
```

Tools:

```text
golangci-lint
gosec
govulncheck
```

---

# Phase 19: Containerization

Docker:

```dockerfile
multi-stage build
```

Add:

```text
docker-compose
```

for local development.

---

# Phase 20: Production Readiness

Create:

```text
Health Checks
Readiness Checks
Liveness Checks
Graceful Shutdown
```

Endpoints:

```text
/health
/ready
/live
```

---

# Phase 21: Documentation

Create:

```text
ARCHITECTURE.md
CONTRIBUTING.md
LOCAL_SETUP.md
DEPLOYMENT.md
```

Document:

* Domains
* Data flow
* Event flow
* Database schema

---

# Final Senior-Level Target Structure

```text
cmd/
internal/
├── domain/
├── application/
├── infrastructure/
├── interfaces/
├── shared/

docs/
deployments/
scripts/
tests/

api/
migrations/
```

---

### AI Prompt You Can Reuse

> You are a Staff-Level Go Backend Engineer. Analyze this ecommerce backend and refactor it incrementally into a production-grade system following Clean Architecture, DDD, Hexagonal Architecture, Go best practices, SOLID principles, OpenTelemetry observability, PostgreSQL optimization, Redis caching, Asynq background jobs, OpenAPI documentation, comprehensive testing, CI/CD, and cloud-native deployment practices. For every change, explain the rationale, impact, migration risk, implementation steps, code examples, and testing strategy. Never perform large rewrites at once; produce a phased migration plan with measurable checkpoints and rollback strategies. 
