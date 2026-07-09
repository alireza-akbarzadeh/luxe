-- +goose Up
-- +goose StatementBegin
INSERT INTO workflow_states (workflow_id, code, name, color, is_initial, is_final, sort_order)
SELECT w.id, t.code, t.name, t.color, FALSE, t.is_final, t.sort_order
FROM workflows w
CROSS JOIN (VALUES
    ('exchange_processing', 'Exchange processing', '#8B5CF6', FALSE, 5),
    ('exchange_completed',  'Exchange completed',  '#10B981', TRUE,  6)
) AS t(code, name, color, is_final, sort_order)
WHERE w.key = 'return'
ON CONFLICT DO NOTHING;

INSERT INTO workflow_transitions (workflow_id, from_state_id, to_state_id, event, name, required_role, guard_key, hook_key)
SELECT w.id, fs.id, ts.id, t.event, t.name, t.required_role, t.guard_key, t.hook_key
FROM workflows w
JOIN (VALUES
    ('item_received', 'exchange_processing', 'start_exchange',    'Start exchange',    'admin', NULL, NULL),
    ('exchange_processing', 'exchange_completed', 'complete_exchange', 'Complete exchange', 'admin', NULL, NULL),
    ('exchange_completed', 'closed', 'close', 'Close exchange', NULL, NULL, NULL)
) AS t(from_code, to_code, event, name, required_role, guard_key, hook_key) ON TRUE
JOIN workflow_states fs ON fs.workflow_id = w.id AND fs.code = t.from_code
JOIN workflow_states ts ON ts.workflow_id = w.id AND ts.code = t.to_code
WHERE w.key = 'return'
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM workflow_transitions
WHERE workflow_id = (SELECT id FROM workflows WHERE key = 'return')
  AND event IN ('start_exchange', 'complete_exchange');

DELETE FROM workflow_states
WHERE workflow_id = (SELECT id FROM workflows WHERE key = 'return')
  AND code IN ('exchange_processing', 'exchange_completed');
-- +goose StatementEnd
