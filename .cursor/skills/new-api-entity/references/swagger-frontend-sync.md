# Swagger contract sync (tell luxe-front)

Load after `make swagger` when the OpenAPI contract changed.

## Required handoff to frontend

After **`make swagger`**, always:

1. **Restart the API** — `/openapi` is cached at process start; restart is mandatory.
2. Tell the user (or agent on **luxe-front**) to run **`pnpm api:gen`** — not optional when the contract changed.

`make swagger` alone does **not** update luxe-front `src/services/`.

## When frontend must regen

| Backend change | Frontend impact |
|----------------|-----------------|
| New `@Router` / handler | New Orval hook file |
| New or renamed DTO | New/renamed `Dto*` in `*.schemas.ts` |
| Renamed JSON fields | TypeScript compile errors until mappers updated |
| Changed `@Success` / response wrapper | Wrong or missing generated types |
| Path or method change | Old hook path invalid |

## Agent reminder text

> Swagger/OpenAPI changed. Restart the luxe API, then in **luxe-front** run `pnpm api:gen` and `pnpm check`. Update domain mappers if DTO field names changed. Do not hand-edit `src/services/`.

Full frontend steps: **luxe-front** skill `/api-gen` → `references/swagger-contract-sync.md`.
