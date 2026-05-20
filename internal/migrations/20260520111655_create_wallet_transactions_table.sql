-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wallet_transactions (
                                                   id             BIGSERIAL PRIMARY KEY,
                                                   created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                                   updated_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                                   deleted_at     TIMESTAMP WITH TIME ZONE,
                                                   user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                                   amount         DECIMAL(10,2) NOT NULL CHECK (amount != 0),
                                                   type           TEXT NOT NULL CHECK (type IN ('deposit', 'payment', 'refund', 'adjustment')),
                                                   reference_type TEXT, -- 'order', 'payment', 'admin'
                                                   reference_id   BIGINT, -- ID of the referenced entity (order.id, payment.id, etc.)
                                                   description    TEXT,
                                                   balance_after  DECIMAL(10,2) NOT NULL,
                                                   status         TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'cancelled')),
                                                   metadata       JSONB
);

CREATE INDEX idx_wallet_transactions_user_id ON wallet_transactions(user_id);
CREATE INDEX idx_wallet_transactions_type ON wallet_transactions(type);
CREATE INDEX idx_wallet_transactions_reference ON wallet_transactions(reference_type, reference_id);
CREATE INDEX idx_wallet_transactions_status ON wallet_transactions(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wallet_transactions;
-- +goose StatementEnd