# Application layer

Use-case orchestration (commands/queries/services). Coordinates domain services and repository ports.

- One package per bounded context (`cart`, `order`, `catalog`, `auth`, `admin`, …)
- No HTTP or Gin imports
- **`apps/wire.go`** — `WireApplications` registers all use cases on `Applications`
- **`bootstrap/wire.go`** — `NewRuntime` wires orchestrators (checkout, inventory, notification, …)

**Pattern:**

```
Handler → apps.Applications → application/<ctx> → infrastructure/postgres
```

**Adding a feature:** see root `README.md` and `.cursor/skills/new-api-entity/SKILL.md`.
