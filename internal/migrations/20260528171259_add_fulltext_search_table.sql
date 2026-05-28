-- +goose Up
-- +goose StatementBegin

-- Enable pg_trgm extension (for trigram similarity, useful for suggestions)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Add tsvector column to products
ALTER TABLE products ADD COLUMN search_vector tsvector 
  GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(name, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(description, '')), 'B')
  ) STORED;

-- Create GIN index on the tsvector
CREATE INDEX idx_products_search ON products USING GIN(search_vector);

-- Add tsvector to stores
ALTER TABLE stores ADD COLUMN search_vector tsvector
  GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(name, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(description, '')), 'B')
  ) STORED;

CREATE INDEX idx_stores_search ON stores USING GIN(search_vector);

-- Add tsvector to categories
ALTER TABLE categories ADD COLUMN search_vector tsvector
  GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(name, '')), 'A')
  ) STORED;

CREATE INDEX idx_categories_search ON categories USING GIN(search_vector);

-- Table for search logs (trending)
CREATE TABLE search_logs (
    id SERIAL PRIMARY KEY,
    query TEXT NOT NULL,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_search_logs_query ON search_logs(query);
CREATE INDEX idx_search_logs_created_at ON search_logs(created_at DESC);
CREATE INDEX idx_search_logs_user_id ON search_logs(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS search_logs;
DROP INDEX IF EXISTS idx_products_search;
DROP INDEX IF EXISTS idx_stores_search;
DROP INDEX IF EXISTS idx_categories_search;
ALTER TABLE products DROP COLUMN IF EXISTS search_vector;
ALTER TABLE stores DROP COLUMN IF EXISTS search_vector;
ALTER TABLE categories DROP COLUMN IF EXISTS search_vector;
DROP EXTENSION IF EXISTS pg_trgm;
-- +goose StatementEnd