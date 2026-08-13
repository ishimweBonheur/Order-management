# Notification Service

Independent Go module that consumes `order.created` and uses Redis event IDs for idempotency.

Set `REDIS_ADDR` and `KAFKA_BROKERS`. Run `go test ./...` and `go run ./cmd/api`.
