-- +goose Up

CREATE INDEX idx_products_category
ON products(category);

CREATE INDEX idx_products_created_at
ON products(created_at DESC);

CREATE INDEX idx_products_price
ON products(price);

CREATE INDEX idx_products_name
ON products(name);

CREATE INDEX idx_products_search
ON products USING GIN (to_tsvector('simple', name || ' ' || description));

-- +goose Down

DROP INDEX IF EXISTS idx_products_category;
DROP INDEX IF EXISTS idx_products_created_at;
DROP INDEX IF EXISTS idx_products_price;
DROP INDEX IF EXISTS idx_products_name;
DROP INDEX IF EXISTS idx_products_search;
