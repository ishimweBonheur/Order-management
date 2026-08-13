# Request walkthrough: POST /orders

```text
Client -> API Gateway -> Auth / Product / Order -> PostgreSQL
                         Product <-> Redis
                         Order -> Kafka -> Notification -> SMTP
                                              |
                                            Redis
```

The client sends its bearer JWT through the gateway. Middleware verifies the HMAC signature and expiry and places user ID and role in context. The handler validates JSON and the service rejects empty, duplicate, or non-positive items. The repository begins a serializable transaction, selects every product `FOR UPDATE`, verifies stock, calculates prices on the server, inserts the order and items, decrements stock, then commits. Any error rolls everything back. After commit the service looks up the sole admin in PostgreSQL and publishes a versioned `order.created` event. Notification reserves the event ID in Redis, sends the admin email, records successful delivery, and then Kafka commits the offset. A redelivered completed event ID is skipped.

Production evolution: use an outbox row written in the same order transaction, then publish it asynchronously. That closes the crash window between database commit and Kafka publication that this learning implementation still has.
