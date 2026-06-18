-- +goose Up
-- +goose StatementBegin

ALTER TABLE brands
    ADD COLUMN IF NOT EXISTS workflow_state_id BIGINT REFERENCES workflow_states(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_brands_workflow_state_id ON brands(workflow_state_id);

INSERT INTO workflows (key, name, description, entity_type) VALUES
    ('brand', 'Brand Lifecycle', 'Brand visibility and catalog lifecycle', 'brand')
ON CONFLICT (key) DO NOTHING;

INSERT INTO workflow_states (workflow_id, code, name, color, is_initial, is_final, sort_order)
SELECT w.id, v.code, v.name, v.color, v.is_initial, v.is_final, v.sort_order
FROM workflows w, (VALUES
    ('draft',    'Draft',    '#9CA3AF', TRUE,  FALSE, 1),
    ('active',   'Active',   '#10B981', FALSE, FALSE, 2),
    ('inactive', 'Inactive', '#6B7280', FALSE, FALSE, 3),
    ('archived', 'Archived', '#374151', FALSE, TRUE,  4)
) AS v(code, name, color, is_initial, is_final, sort_order)
WHERE w.key = 'brand'
ON CONFLICT (workflow_id, code) DO NOTHING;

INSERT INTO workflow_transitions (workflow_id, from_state_id, to_state_id, event, name, required_role, guard_key, hook_key)
SELECT w.id, fs.id, ts.id, t.event, t.name, t.required_role, t.guard_key, t.hook_key
FROM workflows w
JOIN (VALUES
    ('draft',    'active',   'activate',   'Activate',   'admin', NULL, NULL),
    ('active',   'inactive', 'deactivate', 'Deactivate', 'admin', NULL, NULL),
    ('inactive', 'active',   'reactivate', 'Reactivate', 'admin', NULL, NULL),
    ('active',   'archived', 'archive',    'Archive',    'admin', NULL, NULL),
    ('inactive', 'archived', 'archive',    'Archive',    'admin', NULL, NULL),
    ('draft',    'archived', 'archive',    'Archive',    'admin', NULL, NULL)
) AS t(from_code, to_code, event, name, required_role, guard_key, hook_key) ON TRUE
JOIN workflow_states fs ON fs.workflow_id = w.id AND fs.code = t.from_code
JOIN workflow_states ts ON ts.workflow_id = w.id AND ts.code = t.to_code
WHERE w.key = 'brand'
ON CONFLICT DO NOTHING;

UPDATE brands b SET workflow_state_id = s.id
FROM workflow_states s
JOIN workflows w ON w.id = s.workflow_id AND w.key = 'brand'
WHERE b.workflow_state_id IS NULL
  AND s.code = CASE b.status
      WHEN 'draft'    THEN 'draft'
      WHEN 'active'   THEN 'active'
      WHEN 'inactive' THEN 'inactive'
      WHEN 'archived' THEN 'archived'
      ELSE 'draft'
  END;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE brands SET workflow_state_id = NULL;
DELETE FROM workflow_transition_logs WHERE workflow_id IN (SELECT id FROM workflows WHERE key = 'brand');
DELETE FROM workflow_transitions     WHERE workflow_id IN (SELECT id FROM workflows WHERE key = 'brand');
DELETE FROM workflow_states          WHERE workflow_id IN (SELECT id FROM workflows WHERE key = 'brand');
DELETE FROM workflows WHERE key = 'brand';
ALTER TABLE brands DROP COLUMN IF EXISTS workflow_state_id;
-- +goose StatementEnd
