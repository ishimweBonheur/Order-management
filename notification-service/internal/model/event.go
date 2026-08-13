package model

import (
	"encoding/json"
	"time"
)

type Event struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	Version   int             `json:"version"`
	Payload   json.RawMessage `json:"payload"`
}
