package response

// AssetResponse represents asset data in API responses.
type AssetResponse struct {
	ID            string   `json:"id"`
	UserID        string   `json:"user_id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Format        string   `json:"format"`
	FileSize      int64    `json:"file_size"`
	Tags          []string `json:"tags"`
	IsPublic      bool     `json:"is_public"`
	DownloadCount int      `json:"download_count"`
	Status        string   `json:"status"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// PresignedUploadResponse contains the presigned URL for client-side upload.
type PresignedUploadResponse struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
	ExpiresIn int    `json:"expires_in"`
}
