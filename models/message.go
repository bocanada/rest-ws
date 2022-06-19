package models

var (
	PostCreatedMessage = "PostCreated"
	PostUpdatedMessage = "PostUpdated"
	PostDeletedMessage = "PostDeleted"
)

type WebSocketMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}
