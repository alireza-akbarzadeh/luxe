---
name: add-api-endpoint
description: >
  Use when adding a new HTTP handler to an existing luxe Go service — extra action,
  admin route, bulk operation, or method on a resource that already has a model.
  Apply when extending categories, orders, coupons, etc. without a new database table.
  Do not use for new entities/migrations from scratch (use new-api-entity) or
  luxe-front Orval regen alone.
---

# Add API endpoint

**Default:** copy the nearest handler in the same `*_controller.go` / `*_service.go`.

## Checklist

```text
- [ ] DTO in internal/dto/ (if new request/response shapes)
- [ ] Service method — ctx first, db.WithContext(ctx)
- [ ] Controller — bind, validate, utils.Response, no business logic
- [ ] Route in existing *_routes.go
- [ ] Swagger @Router @Success utils.Response{data=…}
- [ ] make swagger
- [ ] Test the path
- [ ] Restart API → luxe-front: `pnpm api:gen` → `pnpm check` (mandatory if DTO/route/Swagger changed)
```

## Gotchas

- **New columns/tables needed?** Stop — run `/new-api-entity` migration first.
- **Multi-table mutation** → `db.Transaction` in the service.
- **`utils.Response{data=dto.X}` in @Success** — bare `dto.X` breaks Orval type generation on the frontend.
- **Constants for statuses/roles** — `internal/constants`.
- **Both repos:** `make swagger` + **restart** + luxe-front **`pnpm api:gen`** whenever the OpenAPI contract changes (new DTO, renamed field, new/changed route). See `new-api-entity/references/swagger-frontend-sync.md`.

## Validate

```bash
go test ./internal/services/... -run TestYourFeature
make swagger && go build ./...
```

Handler template: see [references/handler-pattern.md](../new-api-entity/references/handler-pattern.md) in the `new-api-entity` skill.
