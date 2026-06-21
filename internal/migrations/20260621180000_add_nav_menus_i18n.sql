-- +goose Up
-- +goose StatementBegin
ALTER TABLE nav_menus
    ADD COLUMN IF NOT EXISTS label_i18n JSONB,
    ADD COLUMN IF NOT EXISTS badge_i18n JSONB;

UPDATE nav_menus
SET label_i18n = jsonb_build_object('en', label)
WHERE label_i18n IS NULL AND label <> '';

UPDATE nav_menus
SET badge_i18n = jsonb_build_object('en', badge)
WHERE badge_i18n IS NULL AND badge IS NOT NULL AND badge <> '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE nav_menus
    DROP COLUMN IF EXISTS badge_i18n,
    DROP COLUMN IF EXISTS label_i18n;
-- +goose StatementEnd
