-- +goose Up
-- +goose StatementBegin
ALTER TABLE wallet_transactions
    ADD COLUMN IF NOT EXISTS stripe_session_id VARCHAR(255);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wallet_transactions_stripe_session_id
    ON wallet_transactions(stripe_session_id)
    WHERE stripe_session_id IS NOT NULL AND stripe_session_id <> '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_wallet_transactions_stripe_session_id;

ALTER TABLE wallet_transactions
    DROP COLUMN IF EXISTS stripe_session_id;
-- +goose StatementEnd
