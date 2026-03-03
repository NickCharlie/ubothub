package ws

// MessageType constants define WebSocket message categories.
const (
	TypeChat         = "chat"
	TypeBotReply     = "bot_reply"
	TypeBotStatus    = "bot_status"
	TypeAvatarAction = "avatar_action"
	TypeError        = "error"
	TypePing         = "ping"
	TypePong         = "pong"
)

// InboundMessage represents a message received from a WebSocket client.
type InboundMessage struct {
	Type    string                 `json:"type"`
	BotID   string                 `json:"bot_id,omitempty"`
	Content string                 `json:"content,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// OutboundMessage represents a message sent to a WebSocket client.
type OutboundMessage struct {
	Type      string                 `json:"type"`
	BotID     string                 `json:"bot_id,omitempty"`
	Content   string                 `json:"content,omitempty"`
	Sender    string                 `json:"sender,omitempty"`
	Timestamp int64                  `json:"timestamp,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// RoomMessage carries a message targeting a specific bot room.
type RoomMessage struct {
	RoomID  string
	Message *OutboundMessage
}
