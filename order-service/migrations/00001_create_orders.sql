-- +goose Up
CREATE TABLE orders (
 id UUID PRIMARY KEY,
 user_id UUID NOT NULL REFERENCES users(id),
 status VARCHAR(20) NOT NULL CHECK (status IN ('pending','confirmed','processing','completed','cancelled')),
 total_amount NUMERIC(14,2) NOT NULL CHECK (total_amount >= 0),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE order_items (
 id UUID PRIMARY KEY,
 order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
 product_id UUID NOT NULL REFERENCES products(id),
 quantity INTEGER NOT NULL CHECK (quantity > 0),
 unit_price NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0),
 UNIQUE(order_id,product_id)
);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_user_status ON orders(user_id,status);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
-- +goose Down
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
