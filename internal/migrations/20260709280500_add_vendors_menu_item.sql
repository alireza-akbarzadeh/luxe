-- +goose Up
-- +goose StatementBegin
-- Task 019: Vendors admin hub under Administration.

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Vendors', '/dashboard/vendors', 'BuildingStore', NULL, 7
FROM menu_groups g
JOIN menu_items p ON p.label = 'Administration' AND p.parent_id IS NULL
WHERE g.name = 'Administration'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/vendors');

-- Legacy applications page is folded into the vendors hub pending filter.
DELETE FROM menu_items WHERE href = '/dashboard/vendor-applications';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items WHERE href = '/dashboard/vendors';

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Vendor applications', '/dashboard/vendor-applications', 'UserCheck', NULL, 8
FROM menu_groups g
JOIN menu_items p ON p.label = 'Administration' AND p.parent_id IS NULL
WHERE g.name = 'Administration'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/vendor-applications');
-- +goose StatementEnd
