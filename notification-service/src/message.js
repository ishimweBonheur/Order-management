export function validateOrderCreated(event) {
  if (!event?.event_id || event.event_type !== "order.created" || !event.payload?.admin_email) {
    throw new Error("invalid order.created event or missing admin_email");
  }
  return event;
}

export function createOrderEmail(event, from) {
  validateOrderCreated(event);
  const amount = Number(event.payload.total_amount).toFixed(2);
  const productCount = event.payload.items?.length || 0;
  return {
    from,
    to: event.payload.admin_email,
    subject: `New order ${event.payload.order_id}`,
    text: `A new order ${event.payload.order_id} was created by user ${event.payload.user_id}. Total: ${amount}. Products: ${productCount}.`,
    html: `<h2>New order received</h2><p>Order <strong>${event.payload.order_id}</strong> was created by user ${event.payload.user_id}.</p><p>Total: <strong>${amount}</strong></p><p>Products: <strong>${productCount}</strong></p>`,
  };
}
