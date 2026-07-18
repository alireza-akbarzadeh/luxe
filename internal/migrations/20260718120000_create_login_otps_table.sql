-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS login_otps (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    code_hash VARCHAR(64) NOT NULL,
    channel VARCHAR(16) NOT NULL,
    destination VARCHAR(255) NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_login_otps_user_id ON login_otps(user_id);
CREATE INDEX IF NOT EXISTS idx_login_otps_code_hash ON login_otps(code_hash);

ALTER TABLE login_otps
    DROP CONSTRAINT IF EXISTS fk_login_otps_user;

ALTER TABLE login_otps
    ADD CONSTRAINT fk_login_otps_user
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS login_otps;
-- +goose StatementEnd
