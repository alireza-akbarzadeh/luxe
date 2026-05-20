-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS wallets (
                                       id         BIGSERIAL PRIMARY KEY,
                                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                       updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                       deleted_at TIMESTAMP WITH TIME ZONE,
                                       user_id    BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
                                       balance    DECIMAL(10,2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
                                       version    BIGINT NOT NULL DEFAULT 0,
                                       currency   TEXT NOT NULL DEFAULT 'USD'
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin 
DROP TABLE IF EXISTS wallets;
-- +goose StatementEnd