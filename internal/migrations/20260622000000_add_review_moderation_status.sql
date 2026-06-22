-- +goose Up
-- +goose StatementBegin
ALTER TABLE reviews
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'approved';

CREATE INDEX IF NOT EXISTS idx_reviews_status ON reviews(status);

ALTER TABLE reviews ALTER COLUMN status SET DEFAULT 'pending';

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Reviews', '/dashboard/reviews', 'Star', NULL, 5
FROM menu_groups g
JOIN menu_items p ON p.label = 'Catalog' AND p.parent_id IS NULL
WHERE g.name = 'Catalog'
  AND NOT EXISTS (
    SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/reviews'
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items WHERE href = '/dashboard/reviews';

DROP INDEX IF EXISTS idx_reviews_status;

ALTER TABLE reviews DROP COLUMN IF EXISTS status;
-- +goose StatementEnd
