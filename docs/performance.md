# PostgreSQL performance exercise

Actual recorded results are available in `docs/performance-results.md`. The SQL below remains the reproducible procedure.

Generate data in a disposable database using `generate_series`, run `ANALYZE`, temporarily drop `idx_orders_user_status`, and capture:

```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM orders WHERE user_id = '<existing-user-id>' AND status = 'completed';
```

Recreate `CREATE INDEX idx_orders_user_status ON orders(user_id,status);`, run `ANALYZE orders`, and repeat the identical query. Record planning time, execution time, scan type, rows examined, buffers, dataset size, PostgreSQL version, and hardware. Use `100000` product rows and `500000` order rows. Compare `Execution Time before / Execution Time after`; results depend on selectivity and cache warmth, so do not claim a universal number.
