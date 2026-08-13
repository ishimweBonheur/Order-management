package notification

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ishimweBonheur/order-management/notification-service/internal/email"
	"github.com/ishimweBonheur/order-management/notification-service/internal/model"
)

const processingTTL = 5 * time.Minute
const processedTTL = 7 * 24 * time.Hour

type Service struct {
	redis  *redis.Client
	sender email.Sender
	logger *slog.Logger
}

func New(redisClient *redis.Client, sender email.Sender, logger *slog.Logger) *Service {
	return &Service{redis: redisClient, sender: sender, logger: logger}
}

func Validate(event model.OrderCreatedEvent) error {
	if event.EventID == "" || event.EventType != "order.created" || event.Payload.AdminEmail == "" || event.Payload.OrderID == "" {
		return fmt.Errorf("invalid order.created event or missing admin_email")
	}
	return nil
}

func BuildMessage(event model.OrderCreatedEvent) email.Message {
	p := event.Payload
	text := fmt.Sprintf("A new order %s was created by user %s. Total: %.2f. Products: %d.", p.OrderID, p.UserID, p.TotalAmount, len(p.Items))
	htmlBody := fmt.Sprintf("<h2>New order received</h2><p>Order <strong>%s</strong> was created by user %s.</p><p>Total: <strong>%.2f</strong></p><p>Products: <strong>%d</strong></p>", html.EscapeString(p.OrderID), html.EscapeString(p.UserID), p.TotalAmount, len(p.Items))
	return email.Message{To: p.AdminEmail, Subject: "New order " + p.OrderID, Text: text, HTML: htmlBody}
}

func (s *Service) Handle(ctx context.Context, event model.OrderCreatedEvent) error {
	if err := Validate(event); err != nil {
		return err
	}
	key := "processed_event:" + event.EventID
	reserved, err := s.redis.SetNX(ctx, key, "processing", processingTTL).Result()
	if err != nil {
		return fmt.Errorf("reserve event: %w", err)
	}
	if !reserved {
		s.logger.Info("duplicate event skipped", "event_id", event.EventID)
		return nil
	}
	if err := s.sender.Send(ctx, BuildMessage(event)); err != nil {
		_ = s.redis.Del(context.WithoutCancel(ctx), key).Err()
		return err
	}
	if err := s.redis.Set(ctx, key, "sent", processedTTL).Err(); err != nil {
		return fmt.Errorf("mark event processed: %w", err)
	}
	s.logger.Info("order email sent", "event_id", event.EventID, "to", event.Payload.AdminEmail)
	return nil
}
