# Order Management System

A learning-oriented Go microservice backend with JWT authentication, PostgreSQL transactions and indexes, Redis caching/rate limiting, and Kafka events.

## Run locally

1. Copy `.env.example` to `.env` and replace both secrets.
2. Run `docker compose up --build -d`.
3. Check `http://localhost:8081/health`, ports 8082, 8083, and 8084 as well.
4. Register at `POST :8081/auth/register`, log in at `POST :8081/auth/login`, then send `Authorization: Bearer <access_token>`.

Interactive API documentation is available at `http://localhost:8080` while the Compose stack is running. The source of truth is `docs/openapi.yaml`. Use the Swagger UI **Authorize** button with the access token returned by login.

Kafka topics, messages, consumer groups, and broker details can be inspected in Kafka UI at `http://localhost:8090`.

PostgreSQL initialization runs only when its volume is new. To rerun it, remove the Compose volume intentionally with `docker compose down -v` (this deletes local database data).

## Services and API

| Service | Port | Endpoints |
|---|---:|---|
| Auth | 8081 | `POST /auth/register`, `POST /auth/login`, `GET /auth/me` |
| Product | 8082 | CRUD `/products`; writes require an admin JWT; list supports `page`, `limit`, `category`, `search`, `sort`, `order` |
| Order | 8083 | `POST/GET /orders`, `GET /orders/{id}`, admin list and status routes |
| Notification | 8084 | Kafka `order.created` consumer, Nodemailer SMTP delivery, and `/health` |

New registrations deliberately receive the `customer` role. Promote a development user using SQL: `UPDATE users SET role='admin' WHERE email='you@example.com';` Public role selection would be a privilege-escalation vulnerability.

## Architecture

```text
Client ──JWT──> Auth (users)       Product (products + Redis)
   │                                      ▲
   └──────────────> Order ──transaction───┘
                         │
                         └── order.created ──> Kafka ──> Notification
                                                        └─ Redis event IDs
```

Each service is an independent Git repository and Go module with its own dependencies, configuration, handlers, business logic, repositories/models, infrastructure helpers, tests, Dockerfile, and README. No service imports Go code from the parent folder or another service repository.

The parent `Order-management` directory is not a Git repository. Its Compose file, database bootstrap, and OpenAPI documentation describe and start the complete multi-repository system; they are orchestration assets rather than shared source-code dependencies.

## Important concepts

- A transaction makes order, item, and stock changes all succeed or all roll back. Product rows are locked (`FOR UPDATE`) and the transaction uses serializable isolation to prevent overselling.
- Product cache hits avoid PostgreSQL; misses load PostgreSQL and populate Redis for five minutes. Updates overwrite and deletes invalidate the key.
- Indexes speed reads by maintaining searchable structures, but add storage and write maintenance. A composite `(user_id,status)` index helps filters using the leftmost column; PostgreSQL may prefer a sequential scan for small tables or low-selectivity queries.
- Kafka normally provides at-least-once delivery here: a message is committed after handling. The notification consumer uses Redis `SETNX` by `event_id` so redelivery does not repeat its effect. At-most-once can lose messages; exactly-once needs coordinated transactional processing and does not automatically make external side effects exactly once.

## Testing and query analysis

Run `go test ./...`. See `docs/performance.md` for the reproducible 100k-product/500k-order data and `EXPLAIN (ANALYZE, BUFFERS)` procedure. Actual timings must be recorded on the machine/database used; invented benchmark numbers are intentionally not included.
