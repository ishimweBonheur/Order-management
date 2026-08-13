import http from "node:http";
import process from "node:process";
import Redis from "ioredis";
import { Kafka, logLevel } from "kafkajs";
import nodemailer from "nodemailer";

const required = (name) => {
  const value = process.env[name]?.trim();
  if (!value) throw new Error(`${name} is required`);
  return value;
};

const config = {
  port: Number(process.env.HTTP_PORT || 8084),
  redis: process.env.REDIS_ADDR || "localhost:6379",
  brokers: (process.env.KAFKA_BROKERS || "localhost:9092").split(","),
  topic: process.env.KAFKA_ORDER_TOPIC || "order.created",
  group: process.env.KAFKA_NOTIFICATION_GROUP || "notification-service",
  smtpHost: required("SMTP_HOST"),
  smtpPort: Number(process.env.SMTP_PORT || 587),
  smtpSecure: process.env.SMTP_SECURE === "true",
  smtpUser: required("SMTP_USER"),
  smtpPassword: required("SMTP_PASSWORD"),
  emailFrom: required("EMAIL_FROM"),
};

const [redisHost, redisPort = "6379"] = config.redis.split(":");
const redis = new Redis({ host: redisHost, port: Number(redisPort), maxRetriesPerRequest: null });
const transporter = nodemailer.createTransport({
  host: config.smtpHost,
  port: config.smtpPort,
  secure: config.smtpSecure,
  auth: { user: config.smtpUser, pass: config.smtpPassword },
});
const kafka = new Kafka({ clientId: "notification-service", brokers: config.brokers, logLevel: logLevel.INFO });
const consumer = kafka.consumer({ groupId: config.group });

async function handleEvent(event) {
  if (!event?.event_id || event.event_type !== "order.created" || !event.payload?.customer_email) {
    throw new Error("invalid order.created event or missing customer_email");
  }
  const key = `processed_event:${event.event_id}`;
  if (await redis.exists(key)) {
    console.info(JSON.stringify({ message: "duplicate event skipped", event_id: event.event_id }));
    return;
  }
  const amount = Number(event.payload.total_amount).toFixed(2);
  const delivery = await transporter.sendMail({
    from: config.emailFrom,
    to: event.payload.customer_email,
    subject: `Order ${event.payload.order_id} received`,
    text: `Your order ${event.payload.order_id} was created successfully. Total: ${amount}.`,
    html: `<h2>Order received</h2><p>Your order <strong>${event.payload.order_id}</strong> was created successfully.</p><p>Total: <strong>${amount}</strong></p>`,
  });
  await redis.set(key, "1", "EX", 7 * 24 * 60 * 60, "NX");
  console.info(JSON.stringify({
    message: "order email sent",
    event_id: event.event_id,
    to: event.payload.customer_email,
    message_id: delivery.messageId,
    accepted: delivery.accepted,
    rejected: delivery.rejected,
    smtp_response: delivery.response,
  }));
}

const server = http.createServer(async (req, res) => {
  if (req.url !== "/health") { res.writeHead(404).end(); return; }
  const redisStatus = redis.status === "ready" ? "ok" : "unavailable";
  res.writeHead(redisStatus === "ok" ? 200 : 503, { "Content-Type": "application/json" });
  res.end(JSON.stringify({ status: redisStatus === "ok" ? "ok" : "degraded", redis: redisStatus }));
});

async function start() {
  await redis.ping();
  await transporter.verify();
  await consumer.connect();
  await consumer.subscribe({ topic: config.topic, fromBeginning: false });
  server.listen(config.port, "0.0.0.0", () => console.info(`notification service listening on ${config.port}`));
  await consumer.run({ eachMessage: async ({ message }) => handleEvent(JSON.parse(message.value.toString("utf8"))) });
}

async function shutdown() {
  server.close();
  await consumer.disconnect().catch(() => {});
  await redis.quit().catch(() => {});
  process.exit(0);
}
process.on("SIGINT", shutdown);
process.on("SIGTERM", shutdown);
start().catch((error) => { console.error(error); process.exit(1); });
