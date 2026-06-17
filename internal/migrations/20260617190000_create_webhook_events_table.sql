-- +goose Up
CREATE TABLE IF NOT EXISTS webhook_events (
    id           BIGSERIAL     PRIMARY KEY,
    event_id     TEXT          NOT NULL UNIQUE,
    event_type   TEXT          NOT NULL,
    source       TEXT          NOT NULL DEFAULT 'stripe',
    status       TEXT          NOT NULL DEFAULT 'received'
                               CHECK (status IN ('received', 'processed', 'failed')),
    payload      JSONB         NOT NULL DEFAULT '{}',
    error_msg    TEXT,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_webhook_events_event_type ON webhook_events (event_type);
CREATE INDEX IF NOT EXISTS idx_webhook_events_status     ON webhook_events (status);
CREATE INDEX IF NOT EXISTS idx_webhook_events_created_at ON webhook_events (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS webhook_events;
