# Request walkthrough: POST /orders

The client sends its bearer JWT. Middleware verifies the HMAC signature and expiry and places user ID and role in context. The handler validates JSON and the service rejects empty, duplicate, or non-positive items. The repository begins a serializable transaction, selects every product `FOR UPDATE`, verifies stock, calculates prices on the server, inserts the order and items, decrements stock, then commits. Any error rolls everything back. After commit the service publishes a versioned `order.created` event. Notification consumes it, atomically records the event ID in Redis, logs the notification, and commits the Kafka offset. A redelivered event ID is skipped.

Production evolution: use an outbox row written in the same order transaction, then publish it asynchronously. That closes the crash window between database commit and Kafka publication that this learning implementation still has.
