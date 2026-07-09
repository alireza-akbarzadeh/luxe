-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS admin_notes TEXT,
    ADD COLUMN IF NOT EXISTS customer_segment VARCHAR(50) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_users_customer_segment ON users (customer_segment);
CREATE INDEX IF NOT EXISTS idx_users_membership_tier ON users (membership_tier);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_customer_segment;
DROP INDEX IF EXISTS idx_users_membership_tier;

ALTER TABLE users
    DROP COLUMN IF EXISTS customer_segment,
    DROP COLUMN IF EXISTS admin_notes;
-- +goose StatementEnd
