-- +goose Up
-- +goose StatementBegin
-- Sidebar renders one child level only — use flat hub links under Marketing.

DELETE FROM menu_items
WHERE href IN (
  '/dashboard/promotions/flash-sales',
  '/dashboard/promotions/banners',
  '/dashboard/promotions/campaigns',
  '/dashboard/marketing/subscribers',
  '/dashboard/marketing/templates',
  '/dashboard/marketing/campaigns'
);

DELETE FROM menu_items
WHERE label IN ('Promotions', 'Email marketing')
  AND href IS NULL
  AND parent_id IN (SELECT id FROM menu_items WHERE label = 'Marketing' AND parent_id IS NULL);

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, v.label, v.href, v.icon, NULL, v.display_order
FROM menu_groups g
JOIN menu_items p ON p.label = 'Marketing' AND p.parent_id IS NULL
CROSS JOIN (VALUES
    ('Promotions', '/dashboard/promotions', 'Flame', 2),
    ('Email marketing', '/dashboard/marketing', 'Mail', 3)
) AS v(label, href, icon, display_order)
WHERE g.name = 'Marketing'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = v.href);

UPDATE menu_items SET display_order = 1 WHERE href = '/dashboard/discounts';
UPDATE menu_items SET display_order = 2 WHERE href = '/dashboard/promotions';
UPDATE menu_items SET display_order = 3 WHERE href = '/dashboard/marketing';
UPDATE menu_items SET display_order = 4 WHERE href = '/dashboard/notifications';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items WHERE href IN ('/dashboard/promotions', '/dashboard/marketing');
-- +goose StatementEnd
