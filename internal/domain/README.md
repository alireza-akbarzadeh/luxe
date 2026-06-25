# Domain layer

Pure Go business rules and repository ports. No GORM, Gin, or framework imports.

Each bounded context has (where migrated):

- `entity.go` — aggregate view
- `repository.go` — persistence **port** (interface; implemented in `infrastructure/postgres/{entity}_repository.go`)
- `service.go` — validation and invariants
- `errors.go` — domain errors (optional)

Shared kernel: `internal/domain/shared/` (`money`, `pagination`).
