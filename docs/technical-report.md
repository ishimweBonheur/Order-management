# Technical report

## Architecture

Clients enter through the Nginx API Gateway on port 8000. It routes authentication, product, order, and admin URLs to independently built services. Auth, Product, Order, and Notification are Go services. Notification uses go-mail for SMTP delivery and remains independently configured, containerized, health-checked, and event-driven.

The learning system shares PostgreSQL so an order and its product stock reductions can be committed in one serializable ACID transaction. A strict database-per-service production design would replace this with stock reservations and a saga. PostgreSQL is the source of truth; Redis holds disposable cache, rate-limit, and event-idempotency state.

## Security and APIs

Passwords are bcrypt hashes. JWTs contain user ID and role, are restricted to HS256, and expire. Middleware authenticates requests and role middleware authorizes admin operations. Public registration always creates a customer. PostgreSQL permits only one admin. API errors use a stable JSON envelope and internal database details are not returned.

Swagger at port 8080 is the API deliverable chosen for this project in place of Postman. It targets the gateway. Kafka UI at port 8090 provides development-time topic and consumer inspection.

## Order and event flow

`POST /orders` validates distinct product UUIDs and positive quantities. PostgreSQL starts a serializable transaction, locks product rows with `FOR UPDATE`, calculates trusted prices, inserts the order and every item, reduces stock, and commits or rolls everything back. The service then reads the sole admin email from PostgreSQL and publishes a versioned `order.created` event containing all items.

Notification consumes the event as part of the `notification-service` group, reserves its event ID in Redis, sends the admin email through SMTP, marks the event sent for seven days, and then allows KafkaJS to commit the offset. SMTP errors remove the reservation and fail processing. Kafka therefore provides at-least-once processing and event IDs reduce duplicate side effects.

There remains a crash window between the database commit and Kafka publish. A production deployment should use a transactional outbox. Similarly, an SMTP server accepting a message proves handoff, not inbox placement.

## Operations

Docker Compose starts PostgreSQL, an idempotent migration job, Redis, Kafka, Kafka topic initialization, four application services, the gateway, Swagger, and Kafka UI. All HTTP applications expose `/health`; Compose waits on dependency health or successful initialization. Services handle termination signals and close HTTP/database/cache/Kafka resources.

Structured logs include service, request ID or event ID, method, path, status, and duration where applicable. Secrets live in the ignored root `.env` and are injected through Compose.

## Testing

Tests cover registration/login/JWT, role assignment and its single-admin rule, product CRUD/filtering, authorization, rate limiting, Redis cache hits/misses/TTL/invalidation, order validation, multiple products, ownership, admin access, status validation, and notification event/email construction. Real infrastructure behavior is verified through Compose health checks and Kafka/SMTP logs.

The separate large-data `EXPLAIN ANALYZE` exercise is intentionally not populated in the normal development database. Its reproducible procedure remains in `docs/performance.md`.
