-- +goose Up
-- +goose StatementBegin

-- Default delivery calendar rule toggles (constants, not demo/seed product data).
INSERT INTO delivery_calendar_rules (rule_key, enabled, label, description, sort_order) VALUES
    ('vendor_closed_skip',        TRUE, 'Skip vendor closed days',        'Advance past days the vendor/store is closed per its working schedule.', 1),
    ('national_holiday_skip',     TRUE, 'Skip national holidays',         'Advance past holidays marked as national in scope.',                     2),
    ('regional_holiday_skip',     TRUE, 'Skip regional holidays',         'Advance past holidays scoped to the order''s region.',                    3),
    ('capacity_full_next_day',    TRUE, 'Roll over when at capacity',     'Move to the next working day when max orders/deliveries per day is reached.', 4),
    ('maintenance_block',         TRUE, 'Block maintenance windows',      'Advance past vendor off days flagged as maintenance or emergency close.', 5),
    ('express_ignore_weekend',    FALSE, 'Express ignores weekends',      'Express shipments may process/courier on weekends, bypassing the weekend skip.', 6),
    ('pickup_ignore_delivery_rules', FALSE, 'Pickup ignores delivery rules', 'Pickup orders skip courier availability and shipping-duration steps.', 7)
ON CONFLICT (rule_key) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM delivery_calendar_rules WHERE rule_key IN (
    'vendor_closed_skip',
    'national_holiday_skip',
    'regional_holiday_skip',
    'capacity_full_next_day',
    'maintenance_block',
    'express_ignore_weekend',
    'pickup_ignore_delivery_rules'
);
-- +goose StatementEnd
