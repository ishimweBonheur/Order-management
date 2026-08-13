package notification

import (
	"strings"
	"testing"

	"github.com/ishimweBonheur/order-management/notification-service/internal/model"
)

func testEvent() model.OrderCreatedEvent {
	return model.OrderCreatedEvent{EventID: "event-1", EventType: "order.created", Payload: model.OrderPayload{AdminEmail: "admin@example.com", OrderID: "order-1", UserID: "user-1", TotalAmount: 25, Items: []model.OrderItem{{}, {}}}}
}

func TestValidateOrderCreated(t *testing.T) {
	if err := Validate(testEvent()); err != nil {
		t.Fatalf("valid event rejected: %v", err)
	}
	event := testEvent()
	event.Payload.AdminEmail = ""
	if err := Validate(event); err == nil {
		t.Fatal("event without database-provided admin email was accepted")
	}
}

func TestBuildMessage(t *testing.T) {
	message := BuildMessage(testEvent())
	if message.To != "admin@example.com" {
		t.Fatalf("recipient = %q", message.To)
	}
	if !strings.Contains(message.Text, "25.00") || !strings.Contains(message.Text, "Products: 2") {
		t.Fatalf("unexpected text: %s", message.Text)
	}
}
