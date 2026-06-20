-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_payments_stripe_session_id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_stripe_session_id
    ON payments(stripe_session_id)
    WHERE stripe_session_id IS NOT NULL AND stripe_session_id <> '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_payments_stripe_session_id;

CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_stripe_session_id
    ON payments(stripe_session_id)
    WHERE stripe_session_id IS NOT NULL;
-- +goose StatementEnd
