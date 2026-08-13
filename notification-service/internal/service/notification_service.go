package service

import (
	"context"
	"errors"
	"github.com/ishimweBonheur/order-management/notification-service/internal/model"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
)

type Service struct {
	redis  *redis.Client
	logger *slog.Logger
}

func New(redis *redis.Client, logger *slog.Logger) *Service {
	return &Service{redis: redis, logger: logger}
}
func (s *Service) Handle(ctx context.Context, event model.Event) error {
	if event.EventID == "" || event.EventType != "order.created" {
		return errors.New("invalid event")
	}
	key := "processed_event:" + event.EventID
	fresh, err := s.redis.SetNX(ctx, key, "1", 7*24*time.Hour).Result()
	if err != nil {
		return err
	}
	if !fresh {
		s.logger.Info("duplicate event skipped", "event_id", event.EventID)
		return nil
	}
	s.logger.Info("Order created successfully. Customer notification will be sent.", "event_id", event.EventID, "event_type", event.EventType)
	return nil
}
