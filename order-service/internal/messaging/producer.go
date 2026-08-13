package messaging

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"time"
)

type Event struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Version   int       `json:"version"`
	Payload   any       `json:"payload"`
}
type Publisher interface {
	Publish(context.Context, string, any) error
}
type Producer struct{ writer *kafka.Writer }

func New(brokers []string, topic string) *Producer {
	return &Producer{writer: &kafka.Writer{Addr: kafka.TCP(brokers...), Topic: topic, Balancer: &kafka.LeastBytes{}, RequiredAcks: kafka.RequireAll}}
}
func (p *Producer) Publish(ctx context.Context, eventType string, payload any) error {
	e := Event{EventID: uuid.NewString(), EventType: eventType, Timestamp: time.Now().UTC(), Version: 1, Payload: payload}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{Key: []byte(e.EventID), Value: b})
}
func (p *Producer) Close() error { return p.writer.Close() }
