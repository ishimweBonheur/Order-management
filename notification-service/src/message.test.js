import assert from "node:assert/strict";
import test from "node:test";
import { createOrderEmail, validateOrderCreated } from "./message.js";

const event = { event_id: "event-1", event_type: "order.created", payload: { admin_email: "admin@example.com", order_id: "order-1", user_id: "user-1", total_amount: 25, items: [{}, {}] } };

test("validates an order.created event", () => assert.equal(validateOrderCreated(event), event));
test("rejects an event without the database-provided admin email", () => assert.throws(() => validateOrderCreated({ ...event, payload: {} }), /admin_email/));
test("builds an admin email including total and product count", () => {
  const mail = createOrderEmail(event, "sender@example.com");
  assert.equal(mail.to, "admin@example.com");
  assert.match(mail.text, /25.00/);
  assert.match(mail.text, /Products: 2/);
});
