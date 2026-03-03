package request

// PresignedUploadRequest represents a request for a presigned upload URL.
type PresignedUploadRequest struct {
	Filename string `json:"filename" binding:"required,max=255"`
	Category string `json:"category" binding:"required,oneof=model_3d model_live2d motion texture"`
	FileSize int64  `json:"file_size" binding:"required,min=1"`
}

// CompleteUploadRequest confirms upload completion and creates the asset record.
type CompleteUploadRequest struct {
	FileKey     string `json:"file_key" binding:"required"`
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"omitempty,max=2048"`
	Category    string `json:"category" binding:"required,oneof=model_3d model_live2d motion texture"`
	Format      string `json:"format" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required,min=1"`
	IsPublic    bool   `json:"is_public"`
	Tags        string `json:"tags" binding:"omitempty"`
}

// UpdateAssetRequest represents an asset metadata update.
type UpdateAssetRequest struct {
	Name        string `json:"name" binding:"omitempty,min=1,max=255"`
	Description string `json:"description" binding:"omitempty,max=2048"`
	IsPublic    *bool  `json:"is_public" binding:"omitempty"`
	Tags        string `json:"tags" binding:"omitempty"`
}

// ListAssetRequest represents pagination and filter parameters.
type ListAssetRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Category string `form:"category" binding:"omitempty,oneof=model_3d model_live2d motion texture"`
	Format   string `form:"format" binding:"omitempty"`
	Status   string `form:"status" binding:"omitempty,oneof=processing ready failed"`
	Search   string `form:"search" binding:"omitempty,max=128"`
}
