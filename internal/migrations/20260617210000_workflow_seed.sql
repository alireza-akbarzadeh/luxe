-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- Workflow definitions (reference configuration required for the engine).
-- ============================================================================
INSERT INTO workflows (key, name, description, entity_type) VALUES
    ('product',  'Product Lifecycle',  'Product publishing lifecycle',   'product'),
    ('order',    'Order Lifecycle',    'Order fulfillment lifecycle',    'order'),
    ('shipment', 'Shipment Lifecycle', 'Shipment delivery lifecycle',    'shipment'),
    ('return',   'Return & Refund',    'Return and refund process',      'return'),
    ('user',     'User Account',       'User account lifecycle',         'user')
ON CONFLICT (key) DO NOTHING;

-- ============================================================================
-- States (code, name, color, is_initial, is_final, sort_order). text_color defaults white.
-- ============================================================================
INSERT INTO workflow_states (workflow_id, code, name, color, is_initial, is_final, sort_order)
SELECT w.id, v.code, v.name, v.color, v.is_initial, v.is_final, v.sort_order
FROM workflows w, (VALUES
    ('draft',         'Draft',         '#9CA3AF', TRUE,  FALSE, 1),
    ('under_review',  'Under Review',  '#F59E0B', FALSE, FALSE, 2),
    ('approved',      'Approved',      '#3B82F6', FALSE, FALSE, 3),
    ('published',     'Published',     '#10B981', FALSE, FALSE, 4),
    ('out_of_stock',  'Out of Stock',  '#EF4444', FALSE, FALSE, 5),
    ('discontinued',  'Discontinued',  '#6B7280', FALSE, FALSE, 6),
    ('archived',      'Archived',      '#374151', FALSE, TRUE,  7)
) AS v(code, name, color, is_initial, is_final, sort_order)
WHERE w.key = 'product'
ON CONFLICT (workflow_id, code) DO NOTHING;

INSERT INTO workflow_states (workflow_id, code, name, color, is_initial, is_final, sort_order)
SELECT w.id, v.code, v.name, v.color, v.is_initial, v.is_final, v.sort_order
FROM workflows w, (VALUES
    ('created',         'Created',         '#9CA3AF', TRUE,  FALSE, 1),
    ('pending_payment', 'Pending Payment', '#F59E0B', FALSE, FALSE, 2),
    ('paid',            'Paid',            '#3B82F6', FALSE, FALSE, 3),
    ('processing',      'Processing',      '#6366F1', FALSE, FALSE, 4),
    ('packed',          'Packed',          '#8B5CF6', FALSE, FALSE, 5),
    ('shipped',         'Shipped',         '#0EA5E9', FALSE, FALSE, 6),
    ('delivered',       'Delivered',       '#14B8A6', FALSE, FALSE, 7),
    ('completed',       'Completed',       '#10B981', FALSE, TRUE,  8),
    ('cancelled',       'Cancelled',       '#EF4444', FALSE, TRUE,  9),
    ('refunded',        'Refunded',        '#6B7280', FALSE, TRUE,  10)
) AS v(code, name, color, is_initial, is_final, sort_order)
WHERE w.key = 'order'
ON CONFLICT (workflow_id, code) DO NOTHING;

INSERT INTO workflow_states (workflow_id, code, name, color, is_initial, is_final, sort_order)
SELECT w.id, v.code, v.name, v.color, v.is_initial, v.is_final, v.sort_order
FROM workflows w, (VALUES
    ('pending',          'Pending',          '#9CA3AF', TRUE,  FALSE, 1),
    ('ready_for_pickup', 'Ready for Pickup', '#F59E0B', FALSE, FALSE, 2),
    ('picked_up',        'Picked Up',        '#6366F1', FALSE, FALSE, 3),
    ('in_transit',       'In Transit',       '#0EA5E9', FALSE, FALSE, 4),
    ('out_for_delivery', 'Out for Delivery', '#8B5CF6', FALSE, FALSE, 5),
    ('delivered',        'Delivered',        '#10B981', FALSE, TRUE,  6),
    ('failed_delivery',  'Failed Delivery',  '#EF4444', FALSE, FALSE, 7),
    ('returned',         'Returned',         '#6B7280', FALSE, TRUE,  8)
) AS v(code, name, color, is_initial, is_final, sort_order)
WHERE w.key = 'shipment'
ON CONFLICT (workflow_id, code) DO NOTHING;

INSERT INTO workflow_states (workflow_id, code, name, color, is_initial, is_final, sort_order)
SELECT w.id, v.code, v.name, v.color, v.is_initial, v.is_final, v.sort_order
FROM workflows w, (VALUES
    ('requested',         'Requested',         '#F59E0B', TRUE,  FALSE, 1),
    ('approved',          'Approved',          '#3B82F6', FALSE, FALSE, 2),
    ('rejected',          'Rejected',          '#EF4444', FALSE, TRUE,  3),
    ('item_received',     'Item Received',     '#6366F1', FALSE, FALSE, 4),
    ('refund_processing', 'Refund Processing', '#8B5CF6', FALSE, FALSE, 5),
    ('refunded',          'Refunded',          '#10B981', FALSE, TRUE,  6),
    ('closed',            'Closed',            '#374151', FALSE, TRUE,  7)
) AS v(code, name, color, is_initial, is_final, sort_order)
WHERE w.key = 'return'
ON CONFLICT (workflow_id, code) DO NOTHING;

INSERT INTO workflow_states (workflow_id, code, name, color, is_initial, is_final, sort_order)
SELECT w.id, v.code, v.name, v.color, v.is_initial, v.is_final, v.sort_order
FROM workflows w, (VALUES
    ('registered',                 'Registered',                 '#9CA3AF', TRUE,  FALSE, 1),
    ('email_verification_pending', 'Email Verification Pending', '#F59E0B', FALSE, FALSE, 2),
    ('active',                     'Active',                     '#10B981', FALSE, FALSE, 3),
    ('suspended',                  'Suspended',                  '#F97316', FALSE, FALSE, 4),
    ('blocked',                    'Blocked',                    '#EF4444', FALSE, FALSE, 5),
    ('deleted',                    'Deleted',                    '#374151', FALSE, TRUE,  6)
) AS v(code, name, color, is_initial, is_final, sort_order)
WHERE w.key = 'user'
ON CONFLICT (workflow_id, code) DO NOTHING;

-- ============================================================================
-- Transitions. from=NULL means a wildcard "any state" transition.
-- ============================================================================

-- Helper pattern: resolve workflow + from/to state by code.
-- Product
INSERT INTO workflow_transitions (workflow_id, from_state_id, to_state_id, event, name, required_role, guard_key, hook_key)
SELECT w.id, fs.id, ts.id, t.event, t.name, t.required_role, t.guard_key, t.hook_key
FROM workflows w
JOIN (VALUES
    ('draft',        'under_review', 'submit_for_review', 'Submit for review', NULL,    'product_has_price', NULL),
    ('under_review', 'approved',     'approve',            'Approve',           'admin', NULL,                NULL),
    ('under_review', 'draft',        'reject',             'Reject',            'admin', NULL,                NULL),
    ('approved',     'published',    'publish',            'Publish',           'admin', NULL,                'product_published'),
    ('published',    'out_of_stock', 'mark_out_of_stock',  'Mark out of stock', NULL,    NULL,                NULL),
    ('out_of_stock', 'published',    'restock',            'Restock',           NULL,    NULL,                NULL),
    ('published',    'discontinued', 'discontinue',        'Discontinue',       'admin', NULL,                NULL),
    ('discontinued', 'archived',     'archive',            'Archive',           'admin', NULL,                NULL)
) AS t(from_code, to_code, event, name, required_role, guard_key, hook_key) ON TRUE
JOIN workflow_states fs ON fs.workflow_id = w.id AND fs.code = t.from_code
JOIN workflow_states ts ON ts.workflow_id = w.id AND ts.code = t.to_code
WHERE w.key = 'product'
ON CONFLICT DO NOTHING;

-- Order (non-wildcard)
INSERT INTO workflow_transitions (workflow_id, from_state_id, to_state_id, event, name, required_role, guard_key, hook_key)
SELECT w.id, fs.id, ts.id, t.event, t.name, t.required_role, t.guard_key, t.hook_key
FROM workflows w
JOIN (VALUES
    ('created',         'pending_payment', 'await_payment',    'Await payment',   NULL,    NULL,                      NULL),
    ('pending_payment', 'paid',            'payment_succeeded','Payment received',NULL,    'order_payment_succeeded', 'order_paid'),
    ('paid',            'processing',      'start_processing', 'Start processing',NULL,    NULL,                      NULL),
    ('processing',      'packed',          'pack',             'Pack',            'admin', NULL,                      NULL),
    ('packed',          'shipped',         'ship',             'Ship',            'admin', NULL,                      'order_shipped'),
    ('shipped',         'delivered',       'deliver',          'Deliver',         NULL,    NULL,                      NULL),
    ('delivered',       'completed',       'complete',         'Complete',        NULL,    NULL,                      NULL),
    ('completed',       'refunded',        'refund',           'Refund',          'admin', NULL,                      'order_refunded')
) AS t(from_code, to_code, event, name, required_role, guard_key, hook_key) ON TRUE
JOIN workflow_states fs ON fs.workflow_id = w.id AND fs.code = t.from_code
JOIN workflow_states ts ON ts.workflow_id = w.id AND ts.code = t.to_code
WHERE w.key = 'order'
ON CONFLICT DO NOTHING;

-- Order (wildcard cancel from any state)
INSERT INTO workflow_transitions (workflow_id, from_state_id, to_state_id, event, name, required_role, guard_key, hook_key)
SELECT w.id, NULL, ts.id, 'cancel', 'Cancel order', NULL, 'order_cancellable', 'order_cancelled'
FROM workflows w
JOIN workflow_states ts ON ts.workflow_id = w.id AND ts.code = 'cancelled'
WHERE w.key = 'order'
ON CONFLICT DO NOTHING;

-- Shipment
INSERT INTO workflow_transitions (workflow_id, from_state_id, to_state_id, event, name, required_role, guard_key, hook_key)
SELECT w.id, fs.id, ts.id, t.event, t.name, t.required_role, t.guard_key, t.hook_key
FROM workflows w
JOIN (VALUES
    ('pending',          'ready_for_pickup', 'ready',            'Ready for pickup',  NULL,    NULL, NULL),
    ('ready_for_pickup', 'picked_up',        'pick_up',          'Picked up',         NULL,    NULL, NULL),
    ('picked_up',        'in_transit',       'depart',           'Depart facility',   NULL,    NULL, NULL),
    ('in_transit',       'out_for_delivery', 'out_for_delivery', 'Out for delivery',  NULL,    NULL, NULL),
    ('out_for_delivery', 'delivered',        'deliver',          'Delivered',         NULL,    NULL, 'shipment_delivered'),
    ('out_for_delivery', 'failed_delivery',  'delivery_failed',  'Delivery failed',   NULL,    NULL, NULL),
    ('failed_delivery',  'out_for_delivery', 'retry_delivery',   'Retry delivery',    NULL,    NULL, NULL),
    ('failed_delivery',  'returned',         'return_to_sender', 'Return to sender',  NULL,    NULL, NULL)
) AS t(from_code, to_code, event, name, required_role, guard_key, hook_key) ON TRUE
JOIN workflow_states fs ON fs.workflow_id = w.id AND fs.code = t.from_code
JOIN workflow_states ts ON ts.workflow_id = w.id AND ts.code = t.to_code
WHERE w.key = 'shipment'
ON CONFLICT DO NOTHING;

-- Return
INSERT INTO workflow_transitions (workflow_id, from_state_id, to_state_id, event, name, required_role, guard_key, hook_key)
SELECT w.id, fs.id, ts.id, t.event, t.name, t.required_role, t.guard_key, t.hook_key
FROM workflows w
JOIN (VALUES
    ('requested',         'approved',          'approve',         'Approve return',   'admin', NULL, NULL),
    ('requested',         'rejected',          'reject',          'Reject return',    'admin', NULL, NULL),
    ('approved',          'item_received',     'receive_item',    'Receive item',     'admin', NULL, NULL),
    ('item_received',     'refund_processing', 'start_refund',    'Start refund',     'admin', NULL, NULL),
    ('refund_processing', 'refunded',          'complete_refund', 'Complete refund',  'admin', NULL, 'return_refunded'),
    ('refunded',          'closed',            'close',           'Close',            NULL,    NULL, NULL),
    ('rejected',          'closed',            'close',           'Close',            NULL,    NULL, NULL)
) AS t(from_code, to_code, event, name, required_role, guard_key, hook_key) ON TRUE
JOIN workflow_states fs ON fs.workflow_id = w.id AND fs.code = t.from_code
JOIN workflow_states ts ON ts.workflow_id = w.id AND ts.code = t.to_code
WHERE w.key = 'return'
ON CONFLICT DO NOTHING;

-- User (non-wildcard)
INSERT INTO workflow_transitions (workflow_id, from_state_id, to_state_id, event, name, required_role, guard_key, hook_key)
SELECT w.id, fs.id, ts.id, t.event, t.name, t.required_role, t.guard_key, t.hook_key
FROM workflows w
JOIN (VALUES
    ('registered',                 'email_verification_pending', 'request_verification', 'Request verification', NULL,    NULL, NULL),
    ('email_verification_pending', 'active',                     'verify_email',         'Verify email',         NULL,    NULL, NULL),
    ('active',                     'suspended',                  'suspend',              'Suspend',              'admin', NULL, NULL),
    ('suspended',                  'active',                     'reinstate',            'Reinstate',            'admin', NULL, NULL),
    ('active',                     'blocked',                    'block',                'Block',                'admin', NULL, NULL),
    ('suspended',                  'blocked',                    'block',                'Block',                'admin', NULL, NULL),
    ('blocked',                    'active',                     'unblock',              'Unblock',              'admin', NULL, NULL)
) AS t(from_code, to_code, event, name, required_role, guard_key, hook_key) ON TRUE
JOIN workflow_states fs ON fs.workflow_id = w.id AND fs.code = t.from_code
JOIN workflow_states ts ON ts.workflow_id = w.id AND ts.code = t.to_code
WHERE w.key = 'user'
ON CONFLICT DO NOTHING;

-- User (wildcard delete)
INSERT INTO workflow_transitions (workflow_id, from_state_id, to_state_id, event, name, required_role, guard_key, hook_key)
SELECT w.id, NULL, ts.id, 'delete_account', 'Delete account', NULL, NULL, NULL
FROM workflows w
JOIN workflow_states ts ON ts.workflow_id = w.id AND ts.code = 'deleted'
WHERE w.key = 'user'
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Backfill workflow_state_id pointers from existing status values.
-- The legacy status column is left untouched (engine writes the mirror going forward).
-- ============================================================================
UPDATE orders o SET workflow_state_id = s.id
FROM workflow_states s
JOIN workflows w ON w.id = s.workflow_id AND w.key = 'order'
WHERE o.workflow_state_id IS NULL
  AND s.code = CASE o.status
      WHEN 'pending'        THEN 'pending_payment'
      WHEN 'paid'           THEN 'paid'
      WHEN 'processing'     THEN 'processing'
      WHEN 'shipped'        THEN 'shipped'
      WHEN 'delivered'      THEN 'delivered'
      WHEN 'completed'      THEN 'completed'
      WHEN 'cancelled'      THEN 'cancelled'
      WHEN 'refunded'       THEN 'refunded'
      WHEN 'delayed'        THEN 'processing'
      WHEN 'payment_failed' THEN 'cancelled'
      ELSE 'pending_payment'
  END;

UPDATE products p SET workflow_state_id = s.id
FROM workflow_states s
JOIN workflows w ON w.id = s.workflow_id AND w.key = 'product'
WHERE p.workflow_state_id IS NULL
  AND s.code = CASE p.status
      WHEN 'active'   THEN 'published'
      WHEN 'inactive' THEN 'discontinued'
      WHEN 'archived' THEN 'archived'
      WHEN 'draft'    THEN 'draft'
      ELSE 'published'
  END;

UPDATE shipments sh SET workflow_state_id = s.id
FROM workflow_states s
JOIN workflows w ON w.id = s.workflow_id AND w.key = 'shipment'
WHERE sh.workflow_state_id IS NULL
  AND s.code = CASE sh.status
      WHEN 'pending'    THEN 'pending'
      WHEN 'processing' THEN 'ready_for_pickup'
      WHEN 'shipped'    THEN 'in_transit'
      WHEN 'delivered'  THEN 'delivered'
      WHEN 'cancelled'  THEN 'returned'
      ELSE 'pending'
  END;

UPDATE users u SET workflow_state_id = s.id
FROM workflow_states s
JOIN workflows w ON w.id = s.workflow_id AND w.key = 'user'
WHERE u.workflow_state_id IS NULL
  AND s.code = CASE
      WHEN u.is_active = FALSE                 THEN 'blocked'
      WHEN u.email_verified_at IS NOT NULL      THEN 'active'
      ELSE 'email_verification_pending'
  END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE orders    SET workflow_state_id = NULL;
UPDATE products  SET workflow_state_id = NULL;
UPDATE shipments SET workflow_state_id = NULL;
UPDATE users     SET workflow_state_id = NULL;
DELETE FROM workflow_transition_logs WHERE workflow_id IN (SELECT id FROM workflows);
DELETE FROM workflow_transitions     WHERE workflow_id IN (SELECT id FROM workflows);
DELETE FROM workflow_states          WHERE workflow_id IN (SELECT id FROM workflows);
DELETE FROM workflows WHERE key IN ('product', 'order', 'shipment', 'return', 'user');
-- +goose StatementEnd
