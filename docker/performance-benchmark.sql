\set ON_ERROR_STOP on
\timing on

INSERT INTO users(id,name,email,password_hash,role)
VALUES ('00000000-0000-4000-8000-000000000001','Performance User','performance-test@local.invalid','not-a-login-account','customer')
ON CONFLICT (id) DO NOTHING;

INSERT INTO products(id,name,description,price,stock,category,created_at,updated_at)
SELECT gen_random_uuid(), 'Performance Product ' || n, 'Generated benchmark product', ((n % 10000) + 1) / 100.0, 1000, '__performance__', NOW(), NOW()
FROM generate_series(1, GREATEST(0, 100000 - (SELECT COUNT(*) FROM products WHERE category='__performance__'))) AS n;

INSERT INTO orders(id,user_id,status,total_amount,created_at,updated_at)
SELECT gen_random_uuid(), '00000000-0000-4000-8000-000000000001',
       (ARRAY['pending','confirmed','processing','completed','cancelled'])[(n % 5)+1],
       ((n % 100000) + 1) / 100.0,
       NOW() - ((n % 365) || ' days')::interval,
       NOW()
FROM generate_series(1, GREATEST(0, 500000 - (SELECT COUNT(*) FROM orders WHERE user_id='00000000-0000-4000-8000-000000000001'))) AS n;

ANALYZE products;
ANALYZE orders;

DROP INDEX IF EXISTS idx_orders_user_status;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_orders_status;
ANALYZE orders;

\echo '=== WITHOUT INDEX ==='
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM orders
WHERE user_id='00000000-0000-4000-8000-000000000001'
  AND status='completed';

CREATE INDEX idx_orders_user_status ON orders(user_id,status);
ANALYZE orders;

\echo '=== WITH COMPOSITE INDEX ==='
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM orders
WHERE user_id='00000000-0000-4000-8000-000000000001'
  AND status='completed';

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);

\echo '=== DATASET COUNTS ==='
SELECT COUNT(*) AS generated_products FROM products WHERE category='__performance__';
SELECT COUNT(*) AS generated_orders FROM orders WHERE user_id='00000000-0000-4000-8000-000000000001';
