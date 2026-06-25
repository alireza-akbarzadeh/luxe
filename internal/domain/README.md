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
  inventory/stock.go      stock availability + delta rules
  coupon/coupon.go        eligibility + discount calculation
  wallet/wallet.go        balance + amount rules
```

## Where persistence ports live

| Context | Port location | Postgres impl |
|---------|---------------|-----------------|
| Catalog | `domain/catalog/repository.go` | `postgres/product_repository.go` |
| Order, cart, auth, … | `application/<ctx>/reader.go` or `writer.go` | `postgres/*_repository.go` |

Do **not** add a `domain/*/repository.go` unless `application/` constructors take that interface (catalog is the only case today).

## What to add here

Extract **pure rules** from `application/*/service.go` when you find:

- Calculations (discount, totals, stock deltas)
- Eligibility checks (can cancel, can apply coupon, sufficient balance)
- Invariants with no I/O

Do **not** move orchestration, GORM, DTO mapping, or workflow sync here — those stay in **`application/`**.
