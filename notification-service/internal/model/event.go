package model

import "time"

type OrderCreatedEvent struct {
	EventID   string       `json:"event_id"`
	EventType string       `json:"event_type"`
	Timestamp time.Time    `json:"timestamp"`
	Version   int          `json:"version"`
	Payload   OrderPayload `json:"payload"`
}

type OrderPayload struct {
	OrderID     string      `json:"order_id"`
	UserID      string      `json:"user_id"`
	AdminEmail  string      `json:"admin_email"`
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
	CreatedAt   time.Time   `json:"created_at"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}
