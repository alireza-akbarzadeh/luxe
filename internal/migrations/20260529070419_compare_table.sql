-- +goose Up
-- +goose StatementBegin
CREATE TABLE compare_lists (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(255), -- for guest users (store a random token in cookie/localStorage)
    product_ids INT[] NOT NULL DEFAULT '{}', -- array of product IDs
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_or_session UNIQUE (user_id, session_id)
);

CREATE INDEX idx_compare_lists_user_id ON compare_lists(user_id);
CREATE INDEX idx_compare_lists_session_id ON compare_lists(session_id);

-- Trigger for updated_at
CREATE TRIGGER update_compare_lists_updated_at
    BEFORE UPDATE ON compare_lists
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS compare_lists;
-- +goose StatementEnd
