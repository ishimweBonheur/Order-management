# Notification Service

Node.js service that consumes `order.created` from Kafka, sends order-confirmation email through SMTP with Nodemailer, and stores processed event IDs in Redis for seven-day idempotency.

Copy the root `.env.example` to `.env` and configure `SMTP_HOST`, `SMTP_PORT`, `SMTP_SECURE`, `SMTP_USER`, `SMTP_PASSWORD`, and `EMAIL_FROM`. For Gmail on port 587, use `SMTP_SECURE=false` and a Google App Password. Do not use or commit your normal mailbox password.

Run locally with `npm ci && npm start`, or use the root Docker Compose stack. The health endpoint is `GET :8084/health`. If SMTP delivery fails, Kafka does not commit the message offset and retries it.
