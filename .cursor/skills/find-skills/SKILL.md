---
name: find-skills
description: >
  Use when the user asks to find or install agent skills for backend or full-stack work.
  Checks luxe .cursor/skills/ first (new-api-entity, add-api-endpoint), then luxe-front
  skills and skills.sh via npx skills find. Do not use when the task is already covered
  by a local /new-api-entity or /add-api-endpoint workflow.
---

# Find Skills (Luxe API)

**Default order:** local **luxe** skill → **luxe-front** skill → `AGENTS.md` → ecosystem search.

## Local skills (this repo)

| Need | Skill |
|------|-------|
| New table + full domain | `/new-api-entity` |
| New handler on existing service | `/add-api-endpoint` |

Front follow-up after Swagger: tell agent to run **luxe-front** `/api-gen` (`pnpm api:gen` after API restart).

## Front repo skills

See `luxe-front/.cursor/skills/find-skills/references/luxe-skills-map.md`.

## Ecosystem search

```bash
npx skills find go gin swagger
npx skills find openapi
```

Browse https://skills.sh/ — prefer 1K+ installs and trusted authors.

Install: `npx skills add <owner/repo@skill> -y` (project) or `-g -y` (global).

**Do not install** skills that add repositories layer, skip Swagger, or bypass `utils.Response` patterns.

## Create Luxe skill

Copy pattern from `new-api-entity/SKILL.md` → `.cursor/skills/<name>/` + update README.
