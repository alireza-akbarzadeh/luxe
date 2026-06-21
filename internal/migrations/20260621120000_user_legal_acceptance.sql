-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS terms_accepted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS privacy_accepted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS terms_version TEXT,
    ADD COLUMN IF NOT EXISTS privacy_version TEXT;

COMMENT ON COLUMN users.terms_accepted_at IS 'When the user accepted the terms of service';
COMMENT ON COLUMN users.privacy_accepted_at IS 'When the user accepted the privacy policy';
COMMENT ON COLUMN users.terms_version IS 'Version string of accepted terms document';
COMMENT ON COLUMN users.privacy_version IS 'Version string of accepted privacy document';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN IF EXISTS terms_accepted_at,
    DROP COLUMN IF EXISTS privacy_accepted_at,
    DROP COLUMN IF EXISTS terms_version,
    DROP COLUMN IF EXISTS privacy_version;
-- +goose StatementEnd
