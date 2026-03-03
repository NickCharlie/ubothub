package response

import "time"

// BotResponse represents the bot data in API responses.
type BotResponse struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Framework    string     `json:"framework"`
	Status       string     `json:"status"`
	WebhookURL   string     `json:"webhook_url"`
	Config       string     `json:"config"`
	LastActiveAt *time.Time `json:"last_active_at"`
	CreatedAt    string     `json:"created_at"`
	UpdatedAt    string     `json:"updated_at"`
}

// BotWithTokenResponse includes the access token (shown only on creation).
type BotWithTokenResponse struct {
	BotResponse
	AccessToken string `json:"access_token"`
}
