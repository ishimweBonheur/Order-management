# Notification Service

Go service that consumes `order.created` from Kafka, sends order notifications through SMTP with [`github.com/wneessen/go-mail`](https://github.com/wneessen/go-mail), and stores processed event IDs in Redis for seven-day idempotency.

Copy the root `.env.example` to `.env`. The service requires `HTTP_PORT`, `REDIS_ADDR`, `KAFKA_BROKERS`, `KAFKA_ORDER_TOPIC`, `KAFKA_NOTIFICATION_GROUP`, `SMTP_HOST`, `SMTP_PORT`, `SMTP_SECURE`, `SMTP_USER`, `SMTP_PASSWORD`, and `EMAIL_FROM`. For Gmail on port 587, use `SMTP_SECURE=false` and a Google App Password. Do not use or commit your normal mailbox password.

Run locally with `go run ./cmd/api`, or use the root Docker Compose stack. The health endpoint is `GET :8084/health`. The Kafka offset is committed only after the email is delivered. If SMTP delivery fails, the Redis reservation is released and the event is retried. Duplicate event IDs are skipped.
