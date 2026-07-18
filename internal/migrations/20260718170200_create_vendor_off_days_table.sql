-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS vendor_off_days (
    id         BIGSERIAL PRIMARY KEY,
    vendor_id  BIGINT       NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    title      VARCHAR(255) NOT NULL,
    off_type   VARCHAR(30)  NOT NULL DEFAULT 'vacation',
    start_date TIMESTAMPTZ  NOT NULL,
    end_date   TIMESTAMPTZ  NOT NULL,
    notes      TEXT         NOT NULL DEFAULT '',
    status     VARCHAR(20)  NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_vendor_off_days_off_type CHECK (
        off_type IN ('vacation', 'inventory_count', 'maintenance', 'emergency_close', 'personal_leave')
    ),
    CONSTRAINT chk_vendor_off_days_status CHECK (status IN ('draft', 'published'))
);

CREATE INDEX IF NOT EXISTS idx_vendor_off_days_vendor_id ON vendor_off_days(vendor_id);
CREATE INDEX IF NOT EXISTS idx_vendor_off_days_status ON vendor_off_days(status);
CREATE INDEX IF NOT EXISTS idx_vendor_off_days_start_date ON vendor_off_days(start_date);
CREATE INDEX IF NOT EXISTS idx_vendor_off_days_end_date ON vendor_off_days(end_date);
CREATE INDEX IF NOT EXISTS idx_vendor_off_days_deleted_at ON vendor_off_days(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS vendor_off_days;
-- +goose StatementEnd
