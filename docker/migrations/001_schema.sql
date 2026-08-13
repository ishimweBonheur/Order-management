BEGIN;
CREATE TABLE IF NOT EXISTS schema_migrations(version BIGINT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE IF NOT EXISTS users(id UUID PRIMARY KEY,name VARCHAR(150) NOT NULL,email VARCHAR(320) NOT NULL,password_hash TEXT NOT NULL,role VARCHAR(20) NOT NULL DEFAULT 'customer' CHECK(role IN('admin','customer')),created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(LOWER(email));
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_single_admin ON users ((role)) WHERE role='admin';
CREATE TABLE IF NOT EXISTS products(id UUID PRIMARY KEY,name VARCHAR(255) NOT NULL,description TEXT NOT NULL DEFAULT '',price NUMERIC(12,2) NOT NULL CHECK(price>=0),stock INTEGER NOT NULL CHECK(stock>=0),category VARCHAR(100) NOT NULL,created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_products_search ON products USING GIN(to_tsvector('simple',name||' '||description));
CREATE TABLE IF NOT EXISTS orders(id UUID PRIMARY KEY,user_id UUID NOT NULL REFERENCES users(id),status VARCHAR(20) NOT NULL CHECK(status IN('pending','confirmed','processing','completed','cancelled')),total_amount NUMERIC(14,2) NOT NULL CHECK(total_amount>=0),created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE IF NOT EXISTS order_items(id UUID PRIMARY KEY,order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,product_id UUID NOT NULL REFERENCES products(id),quantity INTEGER NOT NULL CHECK(quantity>0),unit_price NUMERIC(12,2) NOT NULL CHECK(unit_price>=0),UNIQUE(order_id,product_id));
DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conrelid='orders'::regclass AND contype='f'
      AND conkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid='orders'::regclass AND attname='user_id')]::smallint[]
  ) THEN
    ALTER TABLE orders ADD CONSTRAINT fk_orders_user FOREIGN KEY(user_id) REFERENCES users(id);
  END IF;
END $$;
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_user_status ON orders(user_id,status);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
INSERT INTO schema_migrations(version) VALUES(1) ON CONFLICT DO NOTHING;
COMMIT;
