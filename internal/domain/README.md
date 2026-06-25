# Domain layer

Pure Go business rules. **No GORM, Gin, HTTP, or database imports.**

## Layout (one file per context when small)

```text
domain/
  catalog/catalog.go      rules + types
  catalog/repository.go   port used by application/catalog
  order/order.go          CanCancel + Order type
  cart/cart.go            ValidateCheckout + Cart type
  checkout/service.go     checkout validation
  brand/brand.go          ValidateName
  category/category.go    ValidateName
  collection/collection.go ValidateTitle
  workflow/types.go       transition DTOs for application/workflow
  shared/money.go         Money value object
  shared/pagination.go    page params
```

## Where persistence ports live

| Context | Port location | Postgres impl |
|---------|---------------|-----------------|
| Catalog | `domain/catalog/repository.go` | `postgres/product_repository.go` |
| Order, cart, auth, … | `application/<ctx>/reader.go` or `writer.go` | `postgres/*_repository.go` |

Do **not** add a `domain/*/repository.go` unless `application/` constructors take that interface (catalog is the only case today).

## What to add here

- Pure rules: `CanCancel`, `ValidateCreate`, `ValidateCheckout`
- Domain errors and small types used by those rules
- Shared value objects (`Money`)

Keep orchestration, GORM, DTO mapping, and workflow in **`application/`**.
