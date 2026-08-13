package messaging

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/ishimweBonheur/order-management/notification-service/internal/model"
)

type Handler interface {
	Handle(context.Context, model.OrderCreatedEvent) error
}

func Consume(ctx context.Context, reader *kafka.Reader, handler Handler, logger *slog.Logger) error {
	for {
		message, err := reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		var event model.OrderCreatedEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			logger.Error("invalid Kafka message skipped", "error", err, "topic", message.Topic, "partition", message.Partition, "offset", message.Offset)
			if err := reader.CommitMessages(ctx, message); err != nil {
				return err
			}
			continue
		}
		for {
			if err := handler.Handle(ctx, event); err == nil {
				break
			} else {
				logger.Error("event processing failed; retrying", "error", err, "event_id", event.EventID)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(5 * time.Second):
				}
			}
		}
		if err := reader.CommitMessages(ctx, message); err != nil {
			return err
		}
	}
}
