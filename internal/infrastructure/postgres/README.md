# PostgreSQL infrastructure

GORM repository implementations for all bounded contexts.

- One `*_repository.go` per aggregate/table group
- Implements application `Reader`/`Writer` ports and domain repository interfaces where defined
- Shared helpers: `errors.go` (`IsNotFound`), `checkout_repository.go` (transaction-scoped checkout writes)

**Rule:** No GORM in `internal/services/` except checkout (orchestrated multi-service transactions).
