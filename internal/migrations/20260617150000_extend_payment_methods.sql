-- +goose Up
-- +goose StatementBegin
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_method_check;

ALTER TABLE payments ADD CONSTRAINT payments_method_check
    CHECK (method IN (
        'credit_card',
        'debit_card',
        'paypal',
        'gift_card',
        'store_credit',
        'mock',
        'stripe',
        'wallet'
    ));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_method_check;

ALTER TABLE payments ADD CONSTRAINT payments_method_check
    CHECK (method IN (
        'credit_card',
        'debit_card',
        'paypal',
        'gift_card',
        'store_credit'
    ));
-- +goose StatementEnd
