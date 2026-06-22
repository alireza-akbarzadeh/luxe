-- +goose Up
-- +goose StatementBegin

ALTER TABLE reviews
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'approved';

ALTER TABLE reviews
    ADD COLUMN IF NOT EXISTS workflow_state_id BIGINT REFERENCES workflow_states(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_reviews_status ON reviews(status);
CREATE INDEX IF NOT EXISTS idx_reviews_workflow_state_id ON reviews(workflow_state_id);

INSERT INTO workflows (key, name, description, entity_type) VALUES
    ('review', 'Product Review Moderation', 'Customer product review approval lifecycle', 'review')
ON CONFLICT (key) DO NOTHING;

INSERT INTO workflow_states (workflow_id, code, name, color, is_initial, is_final, sort_order)
SELECT w.id, v.code, v.name, v.color, v.is_initial, v.is_final, v.sort_order
FROM workflows w, (VALUES
    ('pending',  'Pending',  '#F59E0B', TRUE,  FALSE, 1),
    ('approved', 'Approved', '#10B981', FALSE, FALSE, 2),
    ('rejected', 'Rejected', '#EF4444', FALSE, TRUE,  3)
) AS v(code, name, color, is_initial, is_final, sort_order)
WHERE w.key = 'review'
ON CONFLICT (workflow_id, code) DO NOTHING;

INSERT INTO workflow_transitions (workflow_id, from_state_id, to_state_id, event, name, required_role, guard_key, hook_key)
SELECT w.id, fs.id, ts.id, t.event, t.name, t.required_role, t.guard_key, t.hook_key
FROM workflows w
JOIN (VALUES
    ('pending', 'approved', 'approve', 'Approve', 'admin', NULL, NULL),
    ('pending', 'rejected', 'reject',  'Reject',  'admin', NULL, NULL)
) AS t(from_code, to_code, event, name, required_role, guard_key, hook_key) ON TRUE
JOIN workflow_states fs ON fs.workflow_id = w.id AND fs.code = t.from_code
JOIN workflow_states ts ON ts.workflow_id = w.id AND ts.code = t.to_code
WHERE w.key = 'review'
ON CONFLICT DO NOTHING;

UPDATE reviews r SET workflow_state_id = s.id, status = s.code
FROM workflow_states s
JOIN workflows w ON w.id = s.workflow_id AND w.key = 'review'
WHERE r.workflow_state_id IS NULL
  AND s.code = 'approved';

ALTER TABLE reviews ALTER COLUMN status SET DEFAULT 'pending';

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Reviews', '/dashboard/reviews', 'Star', NULL, 5
FROM menu_groups g
JOIN menu_items p ON p.label = 'Catalog' AND p.parent_id IS NULL
WHERE g.name = 'Catalog'
  AND NOT EXISTS (
    SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/reviews'
  );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items WHERE href = '/dashboard/reviews';

UPDATE reviews SET workflow_state_id = NULL;

DELETE FROM workflow_transition_logs WHERE workflow_id IN (SELECT id FROM workflows WHERE key = 'review');
DELETE FROM workflow_transitions     WHERE workflow_id IN (SELECT id FROM workflows WHERE key = 'review');
DELETE FROM workflow_states          WHERE workflow_id IN (SELECT id FROM workflows WHERE key = 'review');
DELETE FROM workflows WHERE key = 'review';

DROP INDEX IF EXISTS idx_reviews_workflow_state_id;
DROP INDEX IF EXISTS idx_reviews_status;

ALTER TABLE reviews DROP COLUMN IF EXISTS workflow_state_id;
ALTER TABLE reviews DROP COLUMN IF EXISTS status;
-- +goose StatementEnd
