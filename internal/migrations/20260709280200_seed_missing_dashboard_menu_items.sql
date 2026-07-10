-- +goose Up
-- +goose StatementBegin
-- Add dashboard sidebar links for built admin pages that were missing from the ecommerce menu seed.
-- Idempotent: safe to re-run; skips rows that already exist.

-- Catalog → Collections
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Collections', '/dashboard/collections', 'Library', NULL, 6
FROM menu_groups g
JOIN menu_items p ON p.label = 'Catalog' AND p.parent_id IS NULL
WHERE g.name = 'Catalog'
  AND NOT EXISTS (
    SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/collections'
  );

-- Marketing → Newsletters (hub page still exists; removed in 20260709280000)
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Newsletters', '/dashboard/marketing/newsletters', 'Mail', NULL, 5
FROM menu_groups g
JOIN menu_items p ON p.label = 'Marketing' AND p.parent_id IS NULL
WHERE g.name = 'Marketing'
  AND NOT EXISTS (
    SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/marketing/newsletters'
  );

-- Support group (ticketing module)
INSERT INTO menu_groups (name, display_order)
SELECT 'Support', 8
WHERE NOT EXISTS (SELECT 1 FROM menu_groups WHERE name = 'Support');

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, NULL, 'Support tickets', '/dashboard/support', 'Headphones', NULL, 1
FROM menu_groups g
WHERE g.name = 'Support'
  AND NOT EXISTS (
    SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/support'
  );

-- Administration → stores, shipping providers, settings sub-pages
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, v.label, v.href, v.icon, NULL, v.display_order
FROM menu_groups g
JOIN menu_items p ON p.label = 'Administration' AND p.parent_id IS NULL
CROSS JOIN (VALUES
    ('Stores', '/dashboard/stores', 'BuildingStore', 9),
    ('Shipping providers', '/dashboard/shipping-providers', 'Truck', 10),
    ('Wallet', '/dashboard/settings/wallet', 'CreditCard', 11),
    ('Webhooks', '/dashboard/settings/webhooks', 'Plug', 12),
    ('Audit settings', '/dashboard/settings/audit', 'ShieldCheck', 13)
) AS v(label, href, icon, display_order)
WHERE g.name = 'Administration'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = v.href);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items WHERE href IN (
    '/dashboard/collections',
    '/dashboard/marketing/newsletters',
    '/dashboard/support',
    '/dashboard/stores',
    '/dashboard/shipping-providers',
    '/dashboard/settings/wallet',
    '/dashboard/settings/webhooks',
    '/dashboard/settings/audit'
);

DELETE FROM menu_groups
WHERE name = 'Support'
  AND NOT EXISTS (SELECT 1 FROM menu_items WHERE group_id = menu_groups.id);
-- +goose StatementEnd
