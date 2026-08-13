package messaging

import (
	"context"
	"encoding/json"
	"github.com/ishimweBonheur/order-management/notification-service/internal/model"
	"github.com/ishimweBonheur/order-management/notification-service/internal/service"
	"github.com/segmentio/kafka-go"
	"log/slog"
)

type Consumer struct {
	reader  *kafka.Reader
	service *service.Service
	logger  *slog.Logger
}

func New(brokers []string, topic, group string, s *service.Service, l *slog.Logger) *Consumer {
	return &Consumer{reader: kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, Topic: topic, GroupID: group, MinBytes: 1, MaxBytes: 10e6}), service: s, logger: l}
}
func (c *Consumer) Run(ctx context.Context) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		var event model.Event
		if err = json.Unmarshal(message.Value, &event); err != nil {
			c.logger.Error("invalid Kafka event", "error", err)
			_ = c.reader.CommitMessages(ctx, message)
			continue
		}
		if err = c.service.Handle(ctx, event); err != nil {
			c.logger.Error("event processing failed", "event_id", event.EventID, "error", err)
			continue
		}
		if err = c.reader.CommitMessages(ctx, message); err != nil {
			return err
		}
	}
}
func (c *Consumer) Close() error { return c.reader.Close() }
