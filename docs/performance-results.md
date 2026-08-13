# PostgreSQL EXPLAIN ANALYZE results

Recorded on 2026-08-13 against the local Docker PostgreSQL 17 database. The reproducible generator is `docker/performance-benchmark.sql`.

## Dataset

- Generated products: 100,000
- Generated orders: 500,000
- Benchmark user: `00000000-0000-4000-8000-000000000001`
- Order statuses are evenly distributed; the query returns 100,000 `completed` orders.
- Existing development data was preserved, so the scan encountered 500,014 total order rows at measurement time.

Query:

```sql
SELECT * FROM orders
WHERE user_id='00000000-0000-4000-8000-000000000001'
  AND status='completed';
```

## Without an appropriate index

The single-column and composite order-filter indexes were removed, followed by `ANALYZE orders`.

```text
Seq Scan on orders  (cost=0.00..13530.21 rows=99016 width=63)
                    (actual time=0.350..37.811 rows=100000 loops=1)
  Filter: ((user_id = '00000000-0000-4000-8000-000000000001'::uuid)
           AND ((status)::text = 'completed'::text))
  Rows Removed by Filter: 400014
  Buffers: shared hit=6030
Planning:
  Buffers: shared hit=27
Planning Time: 1.245 ms
Execution Time: 41.285 ms
```

PostgreSQL used a sequential scan and examined every order row, discarding 400,014 rows that did not satisfy both filters.

## With the composite index

Index created:

```sql
CREATE INDEX idx_orders_user_status ON orders(user_id,status);
ANALYZE orders;
```

Plan:

```text
Bitmap Heap Scan on orders  (cost=1384.66..8919.40 rows=100316 width=63)
                            (actual time=3.034..14.942 rows=100000 loops=1)
  Recheck Cond: ((user_id = '00000000-0000-4000-8000-000000000001'::uuid)
                 AND ((status)::text = 'completed'::text))
  Heap Blocks: exact=6030
  Buffers: shared hit=6030 read=90
  -> Bitmap Index Scan on idx_orders_user_status
       (cost=0.00..1359.58 rows=100316 width=0)
       (actual time=2.323..2.324 rows=100000 loops=1)
       Index Cond: ((user_id = '00000000-0000-4000-8000-000000000001'::uuid)
                    AND ((status)::text = 'completed'::text))
       Buffers: shared read=90
Planning:
  Buffers: shared hit=22 read=1
Planning Time: 0.276 ms
Execution Time: 18.729 ms
```

PostgreSQL used the composite index to locate matching tuple positions and then fetched the relevant heap pages. A bitmap scan was appropriate because the result contains 100,000 rows, approximately 20% of the generated orders; individual random index lookups would be inefficient for that many matches.

## Comparison

| Measurement | Without index | With composite index | Change |
|---|---:|---:|---:|
| Planning time | 1.245 ms | 0.276 ms | 77.8% lower |
| Execution time | 41.285 ms | 18.729 ms | 54.6% lower |
| Scan | Sequential | Bitmap index + heap | Targeted matching |

Execution improvement:

```text
41.285 / 18.729 = 2.20x faster
(41.285 - 18.729) / 41.285 × 100 = 54.6% reduction
```

The index improves this query because its column order matches both equality predicates: `(user_id, status)`. The benefit is workload- and cache-dependent. PostgreSQL may still choose a sequential scan for small tables or when most rows match, because reading the table in order can cost less than following an index and then fetching many heap pages. Indexes also consume disk space and make inserts/updates slower because PostgreSQL must maintain them.

After measurement, `idx_orders_user_status`, `idx_orders_user_id`, and `idx_orders_status` were restored for normal application use.
