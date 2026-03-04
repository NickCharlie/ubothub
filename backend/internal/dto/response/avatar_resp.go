package response

// AvatarResponse represents avatar config data in API responses.
type AvatarResponse struct {
	ID            string              `json:"id"`
	UserID        string              `json:"user_id"`
	BotID         *string             `json:"bot_id"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	RenderType    string              `json:"render_type"`
	SceneConfig   string              `json:"scene_config"`
	ActionMapping string              `json:"action_mapping"`
	IsDefault     bool                `json:"is_default"`
	AvatarAssets  []AvatarAssetDetail `json:"avatar_assets,omitempty"`
	CreatedAt     string              `json:"created_at"`
	UpdatedAt     string              `json:"updated_at"`
}

// AvatarAssetDetail represents an avatar-asset binding in API responses.
type AvatarAssetDetail struct {
	AssetID   string `json:"asset_id"`
	AssetName string `json:"asset_name"`
	Role      string `json:"role"`
	Config    string `json:"config"`
	SortOrder int    `json:"sort_order"`
}
