-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS admin_nav_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    favorites JSONB NOT NULL DEFAULT '[]'::jsonb,
    recent JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_nav_preferences_user_id ON admin_nav_preferences(user_id);

ALTER TABLE admin_nav_preferences
    DROP CONSTRAINT IF EXISTS fk_admin_nav_preferences_user;

ALTER TABLE admin_nav_preferences
    ADD CONSTRAINT fk_admin_nav_preferences_user
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS admin_nav_preferences;
-- +goose StatementEnd
