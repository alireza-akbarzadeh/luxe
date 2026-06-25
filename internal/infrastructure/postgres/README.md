# PostgreSQL infrastructure

GORM **repository implementations** for all bounded contexts.

## Why `postgres` and not `repository`?

| Choice | Rationale |
|--------|-----------|
| **`internal/infrastructure/postgres/`** (current) | Infrastructure package names the **technology adapter** (PostgreSQL via GORM). Files inside are already `{entity}_repository.go` — the role is clear from the filename. Matches sibling packages: `asynq/`, `integrations/stripe/`, `workflow/`. |
| ~~`internal/repository/`~~ | Too generic — hides which DB/driver; encourages a flat dumping ground. |
| ~~`internal/packages/...`~~ | Not idiomatic Go; use `internal/` for private app code. |

**Ports vs adapters:** domain/application define **repository interfaces** (ports). This package is the **postgres adapter** that implements them. Do not rename the folder to `repository` without a strong reason — a rename would touch dozens of imports for no functional gain.

## Layout

- One `*_repository.go` per aggregate/table group
- Implements application `Reader`/`Writer` ports and domain repository interfaces where defined
- Shared helpers: `errors.go` (`IsNotFound`), `checkout_repository.go` (transaction-scoped checkout writes)

**Persistence policy:** all database access uses **GORM** against `internal/models/` entities — `Model()`, `Where()`, `Joins()`, `Preload()`, `Transaction()`, etc. Do **not** use `db.Raw()`, `database/sql`, or ad-hoc SQL strings. PostgreSQL-specific expressions (e.g. `date_trunc`, `::date`) belong in GORM `Select`/`Group` clauses on a model, not in raw query strings.

**Rule:** No GORM in handlers. Checkout orchestration may use `db.Transaction` in `application/checkout`.
