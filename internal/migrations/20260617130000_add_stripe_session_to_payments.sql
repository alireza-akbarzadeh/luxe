-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments ADD COLUMN IF NOT EXISTS stripe_session_id VARCHAR(255);

CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_stripe_session_id
    ON payments(stripe_session_id)
    WHERE stripe_session_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_payments_stripe_session_id;
ALTER TABLE payments DROP COLUMN IF EXISTS stripe_session_id;
-- +goose StatementEnd
