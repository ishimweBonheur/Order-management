import http from "node:http";
import process from "node:process";
import Redis from "ioredis";
import { Kafka, logLevel } from "kafkajs";
import nodemailer from "nodemailer";
import { createOrderEmail, validateOrderCreated } from "./message.js";

const required = (name) => {
  const value = process.env[name]?.trim();
  if (!value) throw new Error(`${name} is required`);
  return value;
};

const config = {
  port: Number(required("HTTP_PORT")),
  redis: required("REDIS_ADDR"),
  brokers: required("KAFKA_BROKERS").split(","),
  topic: required("KAFKA_ORDER_TOPIC"),
  group: required("KAFKA_NOTIFICATION_GROUP"),
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
	validateOrderCreated(event);
	const key = `processed_event:${event.event_id}`;
	const reserved = await redis.set(key, "processing", "EX", 300, "NX");
	if (!reserved) {
		console.info(JSON.stringify({ message: "duplicate event skipped", event_id: event.event_id }));
		return;
	}
	let delivery;
	try {
		delivery = await transporter.sendMail(createOrderEmail(event, config.emailFrom));
		await redis.set(key, "sent", "EX", 7 * 24 * 60 * 60);
	} catch (error) {
		await redis.del(key);
		throw error;
	}
  console.info(JSON.stringify({
    message: "order email sent",
    event_id: event.event_id,
	to: event.payload.admin_email,
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
