-- +goose Up
-- +goose StatementBegin

-- Rebuild search_document for rows seeded/updated outside the catalog service (empty index = no matches).
UPDATE products
SET search_document = trim(both FROM concat_ws(' ',
    name,
    description,
    sku,
    barcode,
    slug,
    (SELECT string_agg(value, ' ') FROM jsonb_each_text(COALESCE(name_i18n, '{}'::jsonb))),
    (SELECT string_agg(value, ' ') FROM jsonb_each_text(COALESCE(description_i18n, '{}'::jsonb))),
    array_to_string(tags, ' ')
))
WHERE search_document IS NULL OR btrim(search_document) = '';

UPDATE categories
SET search_document = trim(both FROM concat_ws(' ',
    name,
    description,
    slug,
    (SELECT string_agg(value, ' ') FROM jsonb_each_text(COALESCE(name_i18n, '{}'::jsonb))),
    (SELECT string_agg(value, ' ') FROM jsonb_each_text(COALESCE(description_i18n, '{}'::jsonb)))
))
WHERE search_document IS NULL OR btrim(search_document) = '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Data backfill — no schema rollback.
-- +goose StatementEnd
