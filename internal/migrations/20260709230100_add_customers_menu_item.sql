-- +goose Up
-- +goose StatementBegin
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Customers', '/dashboard/customers', 'UserHeart', NULL, 2
FROM menu_groups g
JOIN menu_items p ON p.label = 'Users' AND p.parent_id IS NOT NULL
WHERE g.name = 'Users & Access'
  AND NOT EXISTS (
    SELECT 1 FROM menu_items mi
    WHERE mi.href = '/dashboard/customers'
  );

UPDATE menu_items
SET display_order = 3
WHERE href = '/dashboard/access-control'
  AND EXISTS (SELECT 1 FROM menu_items WHERE href = '/dashboard/customers');

UPDATE menu_items
SET display_order = 4
WHERE href = '/dashboard/roles'
  AND EXISTS (SELECT 1 FROM menu_items WHERE href = '/dashboard/customers');

UPDATE menu_items
SET display_order = 5
WHERE href = '/dashboard/audit-logs'
  AND EXISTS (SELECT 1 FROM menu_items WHERE href = '/dashboard/customers');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items WHERE href = '/dashboard/customers';

UPDATE menu_items SET display_order = 2 WHERE href = '/dashboard/access-control';
UPDATE menu_items SET display_order = 3 WHERE href = '/dashboard/roles';
UPDATE menu_items SET display_order = 4 WHERE href = '/dashboard/audit-logs';
-- +goose StatementEnd
