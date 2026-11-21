package ws

import "encoding/json"

// EventType identifies a dashboard WebSocket event.
type EventType string

const (
	// NewAgent is broadcast when an agent registers.
	NewAgent   EventType = "agent.New"
	TaskOutput EventType = "task.Output"
	ErrorMsg   EventType = "error.Msg"
)

// WSEvent is the envelope sent to dashboard connections.
type WSEvent struct {
	Event     EventType       `json:"event"`
	MessageID string          `json:"message_id"`
	Body      json.RawMessage `json:"body,omitempty"`
}
