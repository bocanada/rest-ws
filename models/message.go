package models

var (
	PostCreatedMessage = "PostCreated"
	PostUpdatedMessage = "PostUpdated"
)

type WebSocketMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}
