-- +goose Up
-- +goose StatementBegin
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Fulfillment', '/dashboard/fulfillment', 'Package', NULL, 2
FROM menu_groups g
JOIN menu_items p ON p.label = 'Orders & Fulfillment' AND p.parent_id IS NULL
WHERE g.name = 'Orders & Fulfillment'
  AND NOT EXISTS (
    SELECT 1 FROM menu_items mi
    WHERE mi.href = '/dashboard/fulfillment'
  );

UPDATE menu_items
SET display_order = 3
WHERE href = '/dashboard/shipments'
  AND EXISTS (SELECT 1 FROM menu_items WHERE href = '/dashboard/fulfillment');

UPDATE menu_items
SET display_order = 4
WHERE href = '/dashboard/returns'
  AND EXISTS (SELECT 1 FROM menu_items WHERE href = '/dashboard/fulfillment');

UPDATE menu_items
SET display_order = 5
WHERE href = '/dashboard/invoices'
  AND EXISTS (SELECT 1 FROM menu_items WHERE href = '/dashboard/fulfillment');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items WHERE href = '/dashboard/fulfillment';

UPDATE menu_items SET display_order = 2 WHERE href = '/dashboard/shipments';
UPDATE menu_items SET display_order = 3 WHERE href = '/dashboard/returns';
UPDATE menu_items SET display_order = 4 WHERE href = '/dashboard/invoices';
-- +goose StatementEnd
