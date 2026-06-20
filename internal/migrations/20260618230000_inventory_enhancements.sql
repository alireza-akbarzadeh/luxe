-- +goose Up
-- +goose StatementBegin

CREATE UNIQUE INDEX IF NOT EXISTS idx_inventory_adjustments_idempotent_return_product
    ON inventory_adjustments(reference_id, product_id, adjustment_type)
    WHERE reference_type = 'return' AND adjustment_type = 'return_restock';

UPDATE workflow_transitions wt
SET hook_key = 'return_restock_inventory'
FROM workflows w
WHERE wt.workflow_id = w.id
  AND w.key = 'return'
  AND wt.event = 'receive_item';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE workflow_transitions wt
SET hook_key = NULL
FROM workflows w
WHERE wt.workflow_id = w.id
  AND w.key = 'return'
  AND wt.event = 'receive_item';

DROP INDEX IF EXISTS idx_inventory_adjustments_idempotent_return_product;
-- +goose StatementEnd
