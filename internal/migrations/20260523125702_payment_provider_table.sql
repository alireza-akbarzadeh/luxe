-- +goose Up
-- +goose StatementBegin
CREATE TABLE payment_providers (
                                 id            BIGSERIAL PRIMARY KEY,
                                 created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                 updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                 deleted_at    TIMESTAMP WITH TIME ZONE,
                                 name          TEXT NOT NULL UNIQUE CHECK (name IN ('credit_card', 'debit_card', 'paypal', 'gift_card', 'store_credit')),
                                 display_name  TEXT NOT NULL,
                                 description   TEXT,
                                 icon_url      TEXT,
                                 is_active     BOOLEAN NOT NULL DEFAULT true,
                                 sort_order    INT NOT NULL DEFAULT 0,
                                 requires_card BOOLEAN NOT NULL DEFAULT false
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payment_providers;
-- +goose StatementEnd
