-- +goose Up
-- +goose StatementBegin
-- Add provider_id foreign key column
ALTER TABLE shipments ADD COLUMN provider_id INT REFERENCES shipping_providers(id) ON DELETE SET NULL;

-- Add shipping_price column with default 0
ALTER TABLE shipments ADD COLUMN shipping_price DECIMAL(10,2) NOT NULL DEFAULT 0;

-- Create index on provider_id for faster lookups
CREATE INDEX idx_shipments_provider_id ON shipments(provider_id);

-- Optional: backfill shipping_price from carrier (if you have a carriers table with default prices)
-- UPDATE shipments SET shipping_price = COALESCE(
--     (SELECT price FROM shipping_providers WHERE id = shipments.provider_id), 0
-- )
-- WHERE shipping_price = 0 AND provider_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_shipments_provider_id;
ALTER TABLE shipments DROP COLUMN IF EXISTS shipping_price;
ALTER TABLE shipments DROP COLUMN IF EXISTS provider_id;
-- +goose StatementEnd