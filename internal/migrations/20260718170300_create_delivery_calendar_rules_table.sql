-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS delivery_calendar_rules (
    id          BIGSERIAL PRIMARY KEY,
    rule_key    VARCHAR(50)  NOT NULL UNIQUE,
    enabled     BOOLEAN      NOT NULL DEFAULT TRUE,
    label       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_delivery_calendar_rules_sort_order ON delivery_calendar_rules(sort_order);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS delivery_calendar_rules;
-- +goose StatementEnd
