package event

// Event type constants.
const (
	BotMessageReceived  = "bot.message.received"
	BotStatusChanged    = "bot.status.changed"
	AvatarActionTrigger = "avatar.action.trigger"
	AssetUploadComplete = "asset.upload.completed"
	AssetProcessDone    = "asset.process.completed"
)

// BotMessageEvent carries a received bot message.
type BotMessageEvent struct {
	BotID     string                 `json:"bot_id"`
	AvatarID  string                 `json:"avatar_id"`
	Content   string                 `json:"content"`
	Sender    MessageSender          `json:"sender"`
	Group     *MessageGroup          `json:"group,omitempty"`
	Platform  string                 `json:"platform"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp int64                  `json:"timestamp"`
}

// MessageSender represents the sender of a bot message.
type MessageSender struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
}

// MessageGroup represents group info for group messages.
type MessageGroup struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
}

// BotStatusEvent carries a bot status change.
type BotStatusEvent struct {
	BotID  string `json:"bot_id"`
	Status string `json:"status"`
}
