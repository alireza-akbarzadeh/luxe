-- +goose Up
-- +goose StatementBegin
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, NULL, 'Workflow Rules', '/dashboard/workflows', 'GitBranch', NULL,
       COALESCE((SELECT MAX(mi.display_order) FROM menu_items mi WHERE mi.group_id = g.id), 0) + 1
FROM menu_groups g
WHERE g.name IN ('System', 'System Settings')
  AND NOT EXISTS (
    SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/workflows'
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items WHERE href = '/dashboard/workflows';
-- +goose StatementEnd
