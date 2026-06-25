# Application layer

Use-case orchestration (commands/queries). Coordinates domain services and repository ports.

- One package per bounded context (`cart`, `order`, `catalog`, `auth`, `admin`, …)
- No HTTP or Gin imports
- **`bootstrap/wire.go`** — composition root; register new facades here
- Services in `internal/services/` are thin facades delegating here

**Pattern:**

```
Handler → services/*_service.go → application/<ctx> → infrastructure/postgres
```

**Adding a feature:** see root `README.md` and `.cursor/skills/new-api-entity/SKILL.md`.
