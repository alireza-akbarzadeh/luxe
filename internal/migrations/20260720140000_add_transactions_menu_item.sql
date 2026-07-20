-- +goose Up
-- +goose StatementBegin

-- Admin Transactions hub (order payments + wallet ledger), grouped with
-- Invoices under Orders & Fulfillment.
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Transactions', '/dashboard/transactions', 'ArrowsExchange', NULL, 6
FROM menu_groups g
JOIN menu_items p ON p.label = 'Orders & Fulfillment' AND p.parent_id IS NULL
WHERE g.name = 'Orders & Fulfillment'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/transactions');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items WHERE href = '/dashboard/transactions';
-- +goose StatementEnd
