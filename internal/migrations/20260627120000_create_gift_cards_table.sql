-- +goose Up
CREATE TABLE IF NOT EXISTS gift_cards (
    id                BIGSERIAL PRIMARY KEY,
    code              VARCHAR(32) NOT NULL UNIQUE,
    sender_user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    recipient_email   VARCHAR(255) NOT NULL,
    recipient_name    VARCHAR(255),
    sender_name       VARCHAR(255),
    message           TEXT,
    initial_amount    DECIMAL(10, 2) NOT NULL CHECK (initial_amount > 0),
    balance           DECIMAL(10, 2) NOT NULL CHECK (balance >= 0),
    currency          VARCHAR(3) NOT NULL DEFAULT 'USD',
    status            VARCHAR(20) NOT NULL DEFAULT 'active',
    delivery_date     TIMESTAMPTZ,
    expires_at        TIMESTAMPTZ,
    redeemed_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_gift_cards_sender_user_id ON gift_cards(sender_user_id);
CREATE INDEX IF NOT EXISTS idx_gift_cards_recipient_user_id ON gift_cards(recipient_user_id);
CREATE INDEX IF NOT EXISTS idx_gift_cards_recipient_email ON gift_cards(recipient_email);
CREATE INDEX IF NOT EXISTS idx_gift_cards_status ON gift_cards(status);

-- +goose Down
DROP TABLE IF EXISTS gift_cards;
