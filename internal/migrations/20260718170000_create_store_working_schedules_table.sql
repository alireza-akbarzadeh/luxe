-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS store_working_schedules (
    id                       BIGSERIAL PRIMARY KEY,
    store_id                 BIGINT      NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    timezone                 VARCHAR(64) NOT NULL DEFAULT 'UTC',
    working_days             JSONB       NOT NULL DEFAULT '{"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":false,"sun":false}',
    open_time                VARCHAR(5)  NOT NULL DEFAULT '09:00',
    close_time               VARCHAR(5)  NOT NULL DEFAULT '18:00',
    break_start              VARCHAR(5),
    break_end                VARCHAR(5),
    max_orders_per_day       INT         NOT NULL DEFAULT 0,
    max_deliveries_per_day   INT         NOT NULL DEFAULT 0,
    prep_lead_hours          INT         NOT NULL DEFAULT 0,
    processing_lead_hours    INT         NOT NULL DEFAULT 24,
    delivery_buffer_hours    INT         NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMPTZ,
    CONSTRAINT uq_store_working_schedules_store_id UNIQUE (store_id)
);

CREATE INDEX IF NOT EXISTS idx_store_working_schedules_store_id ON store_working_schedules(store_id);
CREATE INDEX IF NOT EXISTS idx_store_working_schedules_deleted_at ON store_working_schedules(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS store_working_schedules;
-- +goose StatementEnd
