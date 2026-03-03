package request

// CreateAvatarRequest represents the avatar creation request payload.
type CreateAvatarRequest struct {
	Name          string `json:"name" binding:"required,min=1,max=128"`
	Description   string `json:"description" binding:"omitempty,max=2048"`
	RenderType    string `json:"render_type" binding:"required,oneof=three_d live2d"`
	SceneConfig   string `json:"scene_config" binding:"omitempty"`
	ActionMapping string `json:"action_mapping" binding:"omitempty"`
}

// UpdateAvatarRequest represents the avatar update request payload.
type UpdateAvatarRequest struct {
	Name          string `json:"name" binding:"omitempty,min=1,max=128"`
	Description   string `json:"description" binding:"omitempty,max=2048"`
	SceneConfig   string `json:"scene_config" binding:"omitempty"`
	ActionMapping string `json:"action_mapping" binding:"omitempty"`
}

// BindBotRequest represents the avatar-bot binding request.
type BindBotRequest struct {
	BotID string `json:"bot_id" binding:"required"`
}

// BindAssetRequest represents the avatar-asset binding request.
type BindAssetRequest struct {
	AssetID   string `json:"asset_id" binding:"required"`
	Role      string `json:"role" binding:"required,oneof=primary_model animation texture accessory"`
	Config    string `json:"config" binding:"omitempty"`
	SortOrder int    `json:"sort_order" binding:"omitempty,min=0"`
}

// UpdateActionMappingRequest represents the action mapping update payload.
type UpdateActionMappingRequest struct {
	ActionMapping string `json:"action_mapping" binding:"required"`
}

// ListAvatarRequest represents pagination parameters for avatar listing.
type ListAvatarRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}
