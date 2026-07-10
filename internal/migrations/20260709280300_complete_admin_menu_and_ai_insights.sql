-- +goose Up
-- +goose StatementBegin
-- Complete admin sidebar for recently added modules + AI business insights (Task 018).
-- Flat links only (sidebar supports one nesting level under group parents).

-- Remove legacy redirect-only newsletters link; email hub is /dashboard/marketing
DELETE FROM menu_items WHERE href = '/dashboard/marketing/newsletters';

-- Catalog → Collections, Reviews
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, v.label, v.href, v.icon, NULL, v.display_order
FROM menu_groups g
JOIN menu_items p ON p.label = 'Catalog' AND p.parent_id IS NULL
CROSS JOIN (VALUES
    ('Collections', '/dashboard/collections', 'Library', 5),
    ('Reviews', '/dashboard/reviews', 'Star', 6)
) AS v(label, href, icon, display_order)
WHERE g.name = 'Catalog'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = v.href);

-- Marketing hubs (promotions, email marketing)
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

UPDATE menu_items SET label = 'Email marketing', display_order = 3
WHERE href = '/dashboard/marketing';
UPDATE menu_items SET label = 'Promotions', display_order = 2
WHERE href = '/dashboard/promotions';
UPDATE menu_items SET display_order = 1 WHERE href = '/dashboard/discounts';
UPDATE menu_items SET display_order = 4 WHERE href = '/dashboard/notifications';

-- Support group
INSERT INTO menu_groups (name, display_order)
SELECT 'Support', 8
WHERE NOT EXISTS (SELECT 1 FROM menu_groups WHERE name = 'Support');

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, NULL, 'Support tickets', '/dashboard/support', 'Headphones', NULL, 1
FROM menu_groups g
WHERE g.name = 'Support'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/support');

-- Reports → Sales analytics + AI business insights
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, v.label, v.href, v.icon, NULL, v.display_order
FROM menu_groups g
JOIN menu_items p ON p.label = 'Reports' AND p.parent_id IS NULL
CROSS JOIN (VALUES
    ('Sales analytics', '/dashboard/analytics', 'ChartBar', 1),
    ('AI business insights', '/dashboard/insights', 'Sparkles', 2),
    ('Revenue report', '/dashboard/reports/revenue', 'TrendingUp', 3),
    ('Traffic report', '/dashboard/reports/traffic', 'DeviceAnalytics', 4)
) AS v(label, href, icon, display_order)
WHERE g.name = 'Reports'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = v.href);

UPDATE menu_items SET label = 'Sales analytics', display_order = 1 WHERE href = '/dashboard/analytics';
UPDATE menu_items SET label = 'AI business insights', display_order = 2 WHERE href = '/dashboard/insights';
UPDATE menu_items SET label = 'Revenue report', display_order = 3 WHERE href = '/dashboard/reports/revenue';
UPDATE menu_items SET label = 'Traffic report', display_order = 4 WHERE href = '/dashboard/reports/traffic';

-- Administration extras
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
    '/dashboard/reviews',
    '/dashboard/insights',
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
