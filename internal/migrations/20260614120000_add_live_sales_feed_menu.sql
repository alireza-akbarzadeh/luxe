-- +goose Up
-- +goose StatementBegin
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, NULL, 'Live Sales Feed', '/dashboard/live', 'Activity', NULL, 2
FROM menu_groups g
WHERE g.name = 'Overview'
  AND NOT EXISTS (
    SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/live'
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items WHERE href = '/dashboard/live';
-- +goose StatementEnd
