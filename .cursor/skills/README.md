# Skill description eval queries

See `eval-queries.json` for train/validation prompts testing `new-api-entity` vs `add-api-endpoint` boundaries.

Manual check: run should-trigger queries in Cursor Agent and confirm the right skill loads. Revise `description` in each `SKILL.md` using the train set; validate on the held-out split before committing.

Descriptions must stay under 1024 characters.
