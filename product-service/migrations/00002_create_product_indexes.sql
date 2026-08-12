-- +goose Up

CREATE INDEX idx_products_category
ON products(category);

CREATE INDEX idx_products_created_at
ON products(created_at DESC);

CREATE INDEX idx_products_price
ON products(price);

-- +goose Down

DROP INDEX IF EXISTS idx_products_category;
DROP INDEX IF EXISTS idx_products_created_at;
DROP INDEX IF EXISTS idx_products_price;