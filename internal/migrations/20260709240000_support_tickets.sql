-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS support_tickets (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id BIGINT REFERENCES users(id),
    customer_name VARCHAR(200) NOT NULL DEFAULT '',
    customer_email VARCHAR(255) NOT NULL DEFAULT '',
    order_id BIGINT REFERENCES orders(id),
    subject VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    priority VARCHAR(32) NOT NULL DEFAULT 'normal',
    channel VARCHAR(32) NOT NULL DEFAULT 'web',
    assignee_id BIGINT REFERENCES users(id),
    admin_notes TEXT,
    last_message_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_support_tickets_user_id ON support_tickets (user_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_status ON support_tickets (status);
CREATE INDEX IF NOT EXISTS idx_support_tickets_channel ON support_tickets (channel);
CREATE INDEX IF NOT EXISTS idx_support_tickets_assignee_id ON support_tickets (assignee_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_last_message_at ON support_tickets (last_message_at DESC);

CREATE TABLE IF NOT EXISTS support_ticket_messages (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ticket_id BIGINT NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    author_id BIGINT REFERENCES users(id),
    author_role VARCHAR(32) NOT NULL DEFAULT 'customer',
    body TEXT NOT NULL,
    channel VARCHAR(32) NOT NULL DEFAULT 'web',
    is_internal BOOLEAN NOT NULL DEFAULT FALSE,
    is_ai_suggestion BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_support_ticket_messages_ticket_id ON support_ticket_messages (ticket_id);
CREATE INDEX IF NOT EXISTS idx_support_ticket_messages_created_at ON support_ticket_messages (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS support_ticket_messages;
DROP TABLE IF EXISTS support_tickets;
-- +goose StatementEnd
