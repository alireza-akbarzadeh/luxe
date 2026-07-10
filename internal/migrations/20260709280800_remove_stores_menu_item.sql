-- +goose Up
-- +goose StatementBegin
-- Consolidate marketplace seller admin under Vendors; Stores menu was duplicate UX.

DELETE FROM menu_items WHERE href = '/dashboard/stores';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Stores', '/dashboard/stores', 'BuildingStore', NULL, 9
FROM menu_groups g
JOIN menu_items p ON p.label = 'Administration' AND p.parent_id IS NULL
WHERE g.name = 'Administration'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/stores');
-- +goose StatementEnd
