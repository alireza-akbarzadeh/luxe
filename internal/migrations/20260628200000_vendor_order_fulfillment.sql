-- +goose Up
-- Allow store owners (non-admin) to pack and ship orders via vendor panel.
-- Authorization is enforced at the HTTP layer (store ownership), not by workflow role.
UPDATE workflow_transitions wt
SET required_role = NULL
FROM workflows w
WHERE wt.workflow_id = w.id
  AND w.key = 'order'
  AND wt.event IN ('pack', 'ship');

-- +goose Down
UPDATE workflow_transitions wt
SET required_role = 'admin'
FROM workflows w
WHERE wt.workflow_id = w.id
  AND w.key = 'order'
  AND wt.event IN ('pack', 'ship');
