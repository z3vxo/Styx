package ws

import "encoding/json"

type EventType string

const (
	TaskCreate     EventType = "task.Create"
	TaskCreateResp EventType = "task.Create.Response"
	TaskOutput     EventType = "task.Output"
	WsError        EventType = "ws.Error"
)

// --------------core--------------
type WSEvent struct {
	Event string          `json:"event"`
	Meta  MetaData        `json:"metadata"`
	Body  json.RawMessage `json:"body,omitempty"`
}

type MetaData struct {
	MessageID string `json:"message_id"`
	TabID     string `json:"tab_id"`
}

type ErrorResp struct {
	Message string `json:"error"`
}

//--------------

// --------------TASKS--------------

type WsNewTaskResponse struct {
	Ok     bool  `json:"ok"`
	TaskID int32 `json:"task_id"`
}

//--------------
