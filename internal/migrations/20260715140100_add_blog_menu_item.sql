-- +goose Up
-- +goose StatementBegin

-- Content group for editorial surfaces
INSERT INTO menu_groups (name, display_order)
SELECT 'Content', 5
WHERE NOT EXISTS (SELECT 1 FROM menu_groups WHERE name = 'Content');

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, NULL, 'Blog', '/dashboard/blog', 'News', NULL, 1
FROM menu_groups g
WHERE g.name = 'Content'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/blog');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items WHERE href = '/dashboard/blog';
DELETE FROM menu_groups WHERE name = 'Content'
  AND NOT EXISTS (SELECT 1 FROM menu_items WHERE group_id = menu_groups.id);
-- +goose StatementEnd
