# Order Service

Independent Go module for customer/admin order APIs, transactional stock updates, and `order.created` Kafka events.

Set the required `HTTP_PORT`, `DATABASE_URL`, `JWT_SECRET`, `KAFKA_BROKERS`, and `KAFKA_ORDER_TOPIC` environment variables. Run `go test ./...` and `go run ./cmd/api`.
