-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS store_holidays (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    description   TEXT         NOT NULL DEFAULT '',
    holiday_type  VARCHAR(20)  NOT NULL DEFAULT 'store',
    start_date    TIMESTAMPTZ  NOT NULL,
    end_date      TIMESTAMPTZ  NOT NULL,
    is_recurring  BOOLEAN      NOT NULL DEFAULT FALSE,
    recurrence_rule VARCHAR(50),
    apply_to      VARCHAR(20)  NOT NULL DEFAULT 'stores',
    vendor_id     BIGINT       REFERENCES stores(id) ON DELETE CASCADE,
    region        VARCHAR(100),
    priority      INT          NOT NULL DEFAULT 0,
    status        VARCHAR(20)  NOT NULL DEFAULT 'draft',
    notes         TEXT         NOT NULL DEFAULT '',
    created_by    BIGINT       REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,
    CONSTRAINT chk_store_holidays_holiday_type CHECK (holiday_type IN ('national', 'regional', 'store', 'vendor')),
    CONSTRAINT chk_store_holidays_apply_to CHECK (apply_to IN ('all', 'stores', 'vendor', 'region')),
    CONSTRAINT chk_store_holidays_status CHECK (status IN ('draft', 'published'))
);

CREATE INDEX IF NOT EXISTS idx_store_holidays_holiday_type ON store_holidays(holiday_type);
CREATE INDEX IF NOT EXISTS idx_store_holidays_status ON store_holidays(status);
CREATE INDEX IF NOT EXISTS idx_store_holidays_start_date ON store_holidays(start_date);
CREATE INDEX IF NOT EXISTS idx_store_holidays_end_date ON store_holidays(end_date);
CREATE INDEX IF NOT EXISTS idx_store_holidays_vendor_id ON store_holidays(vendor_id);
CREATE INDEX IF NOT EXISTS idx_store_holidays_region ON store_holidays(region);
CREATE INDEX IF NOT EXISTS idx_store_holidays_deleted_at ON store_holidays(deleted_at);

-- Junction table: which stores a holiday applies to when apply_to = 'stores'.
CREATE TABLE IF NOT EXISTS store_holiday_stores (
    holiday_id BIGINT NOT NULL REFERENCES store_holidays(id) ON DELETE CASCADE,
    store_id   BIGINT NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    PRIMARY KEY (holiday_id, store_id)
);

CREATE INDEX IF NOT EXISTS idx_store_holiday_stores_store_id ON store_holiday_stores(store_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS store_holiday_stores;
DROP TABLE IF EXISTS store_holidays;
-- +goose StatementEnd
