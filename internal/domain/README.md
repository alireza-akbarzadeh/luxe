# Domain layer

Pure Go business rules and repository ports. No GORM, Gin, or framework imports.

Each bounded context has (where migrated):

- `entity.go` — aggregate view
- `repository.go` — persistence port (optional)
- `service.go` — validation and invariants
- `errors.go` — domain errors (optional)

Shared kernel: `internal/domain/shared/` (`money`, `pagination`).
