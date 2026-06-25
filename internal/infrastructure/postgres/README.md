# PostgreSQL infrastructure

GORM repository implementations for all bounded contexts.

- One `*_repository.go` per aggregate/table group
- Implements application `Reader`/`Writer` ports and domain repository interfaces where defined
- Shared helpers: `errors.go` (`IsNotFound`), `checkout_repository.go` (transaction-scoped checkout writes)

**Persistence policy:** all database access uses **GORM** against `internal/models/` entities — `Model()`, `Where()`, `Joins()`, `Preload()`, `Transaction()`, etc. Do **not** use `db.Raw()`, `database/sql`, or ad-hoc SQL strings. PostgreSQL-specific expressions (e.g. `date_trunc`, `::date`) belong in GORM `Select`/`Group` clauses on a model, not in raw query strings.

**Rule:** No GORM in handlers. Checkout orchestration may use `db.Transaction` in `application/checkout`.
