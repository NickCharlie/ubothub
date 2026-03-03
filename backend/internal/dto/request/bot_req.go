package request

// CreateBotRequest represents the bot creation request payload.
type CreateBotRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Description string `json:"description" binding:"omitempty,max=2048"`
	Framework   string `json:"framework" binding:"required,oneof=astrbot nonebot wechaty koishi custom"`
	WebhookURL  string `json:"webhook_url" binding:"omitempty,url,max=2048"`
	Config      string `json:"config" binding:"omitempty"`
}

// UpdateBotRequest represents the bot update request payload.
type UpdateBotRequest struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=128"`
	Description string `json:"description" binding:"omitempty,max=2048"`
	WebhookURL  string `json:"webhook_url" binding:"omitempty,url,max=2048"`
	Config      string `json:"config" binding:"omitempty"`
}

// ListBotRequest represents pagination parameters for bot listing.
type ListBotRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}
