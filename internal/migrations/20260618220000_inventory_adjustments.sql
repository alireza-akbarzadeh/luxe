-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS inventory_adjustments (
    id              BIGSERIAL PRIMARY KEY,
    product_id      BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity_delta  INT NOT NULL,
    quantity_before INT NOT NULL,
    quantity_after  INT NOT NULL CHECK (quantity_after >= 0),
    adjustment_type VARCHAR(32) NOT NULL,
    reference_type  VARCHAR(32),
    reference_id    BIGINT,
    actor_user_id   BIGINT REFERENCES users(id) ON DELETE SET NULL,
    note            TEXT NOT NULL DEFAULT '',
    metadata        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_inventory_adjustments_product_created
    ON inventory_adjustments(product_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_inventory_adjustments_reference
    ON inventory_adjustments(reference_type, reference_id)
    WHERE reference_type IS NOT NULL AND reference_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_inventory_adjustments_type_created
    ON inventory_adjustments(adjustment_type, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_inventory_adjustments_idempotent_order_product
    ON inventory_adjustments(reference_id, product_id, adjustment_type)
    WHERE reference_type = 'order' AND adjustment_type IN ('sale', 'order_cancel');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS inventory_adjustments;
-- +goose StatementEnd
