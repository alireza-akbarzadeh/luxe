-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE nav_menus (
                           id          SERIAL PRIMARY KEY,
                           label       VARCHAR(100) NOT NULL,
                           type        VARCHAR(20) NOT NULL,
                           href        VARCHAR(500),
                           badge       VARCHAR(50),
                           view_all    JSONB,
                           columns     JSONB,
                           featured    JSONB,
                           "order"     INT DEFAULT 0,
                           created_at  TIMESTAMP DEFAULT NOW(),
                           updated_at  TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_nav_menus_order ON nav_menus("order");

-- +goose Down
-- +goose StatementBegin
    DROP TABLE nav_menus;
-- +goose StatementEnd
