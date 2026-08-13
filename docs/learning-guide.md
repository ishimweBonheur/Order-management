# Learning guide

## Read the project in this order

1. Start at each `cmd/api/main.go`. This is the composition root: configuration and infrastructure are created, dependencies are connected, routes are declared, and shutdown is handled.
2. Read a handler. A handler translates HTTP (JSON, URL parameters, status codes) and should not contain database rules.
3. Read its service. This is where validation and business policy live—for example, public registration always creates a customer and customers cannot read another user's order.
4. Read the repository. It hides SQL and makes business logic testable with an in-memory or fake implementation.
5. Read models, migrations, cache, and messaging adapters. They define stored data and communication boundaries.

## Authentication flow

Registration normalizes the email, validates input, hashes the password with bcrypt, and stores only the hash. Login fetches the user and compares the hash without revealing whether the email or password was wrong. The signed JWT contains user ID, role, issuer, issue time, and expiry. Middleware accepts only HS256, verifies the signature and expiry, and adds identity to request context. Authorization then checks the context role; authentication answers “who are you?” while authorization answers “may you do this?”.

## Product flow

Admin writes pass JWT and role middleware. Reads are public. `GET /products/{id}` checks `product:<uuid>` in Redis; a miss loads PostgreSQL and caches the JSON for five minutes. Update replaces the cached value and delete removes it. List SQL uses parameter placeholders for user values and a fixed allow-list for sort columns, preventing SQL injection. Pagination uses `LIMIT` and `OFFSET`.

## Order flow and transaction

The client never supplies price, total, user ID, or status. User ID comes from the verified JWT; prices and stock come from PostgreSQL. A serializable transaction locks selected product rows. If every item is valid, it inserts the order/items and decrements stock before commit. A failure triggers rollback, so partial orders and lost stock cannot remain. Locking matters because two buyers may request the last unit concurrently.

This exercise keeps products and orders in one PostgreSQL database so that stock and order writes can share one ACID transaction. A stricter database-per-service architecture cannot perform that local transaction across services; it would use reservations plus a saga/compensation workflow.

## Kafka reliability

The producer waits for acknowledgements from all in-sync replicas. The consumer fetches a message, processes it, and only then commits the offset: at-least-once delivery. Redis `SETNX` records each event ID, making the log action idempotent for seven days. A real notification database should enforce a unique event ID permanently. The remaining producer crash window is described in `architecture.md`; implement an outbox before calling this production-grade delivery.

## Indexes

An index is an additional sorted/searchable structure pointing to table rows. It avoids examining every row for selective queries. Every insert/update/delete must also maintain applicable indexes, so unnecessary indexes slow writes and consume space. `(user_id,status)` serves `user_id` and `user_id + status` lookups due to the leftmost-prefix rule, but generally not `status` alone. PostgreSQL can choose a sequential scan when a table is small, many rows match, statistics are stale, or the indexed expression/type differs from the query.

`EXPLAIN` estimates without executing. `EXPLAIN ANALYZE` actually runs the query, so do not use it casually on destructive statements. Compare scan type, estimated versus actual rows, buffers, planning time, and execution time—not execution time alone.

## Suggested hands-on sequence

1. Run all unit tests and read one test beside its service.
2. Start Compose and register/login a customer.
3. Promote that user to admin in the development database and create two products.
4. Log in as a second customer, list products twice, and inspect Redis keys to see caching.
5. Create an order and inspect the order, item, and reduced product stock rows.
6. Watch notification logs, then replay the same event ID and observe the duplicate skip.
7. Run the performance exercise and write down your own query plans and timings.
