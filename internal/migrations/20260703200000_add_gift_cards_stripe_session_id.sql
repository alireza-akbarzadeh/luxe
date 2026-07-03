-- +goose Up
ALTER TABLE gift_cards
    ADD COLUMN IF NOT EXISTS stripe_session_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_gift_cards_stripe_session_id
    ON gift_cards (stripe_session_id)
    WHERE stripe_session_id IS NOT NULL AND stripe_session_id <> '';

-- +goose Down
DROP INDEX IF EXISTS idx_gift_cards_stripe_session_id;
ALTER TABLE gift_cards DROP COLUMN IF EXISTS stripe_session_id;
