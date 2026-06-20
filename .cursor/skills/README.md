# Luxe API — Agent Skills

| Skill | When |
|-------|------|
| `new-api-entity` | New table + full domain (migration → Swagger) |
| `add-api-endpoint` | New handler on an existing service |

Each skill includes `evals/evals.json` for output-quality testing. Root `eval-queries.json` tests description triggering.

## Output quality evals

1. Read `<skill>/evals/evals.json`
2. Run each **prompt** in Agent with `/skill-name`
3. Save outputs under `<skill>-workspace/iteration-1/<eval-name>/with_skill/outputs/`
4. Baseline: same prompt without skill → `without_skill/outputs/`
5. Grade **assertions** in `grading.json`; iterate on `SKILL.md`

See [evaluating skill output quality](https://agentskills.io/skill-creation/evaluating-skills).

## Description triggering

Use `eval-queries.json` train/validation splits. Revise `description` in frontmatter only.
