-- +goose Up
-- +goose StatementBegin
ALTER TABLE coupons
    ADD COLUMN IF NOT EXISTS application_type VARCHAR(32) NOT NULL DEFAULT 'code',
    ADD COLUMN IF NOT EXISTS conditions JSONB NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS bogo_buy_quantity INT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS bogo_get_quantity INT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS bogo_get_discount_percent DECIMAL(5,2) NOT NULL DEFAULT 100;

CREATE INDEX IF NOT EXISTS idx_coupons_application_type ON coupons (application_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_coupons_application_type;
ALTER TABLE coupons
    DROP COLUMN IF EXISTS bogo_get_discount_percent,
    DROP COLUMN IF EXISTS bogo_get_quantity,
    DROP COLUMN IF EXISTS bogo_buy_quantity,
    DROP COLUMN IF EXISTS conditions,
    DROP COLUMN IF EXISTS application_type;
-- +goose StatementEnd
