package dto

// PresignUploadRequest requests a signed upload URL for direct client → R2 upload.
type PresignUploadRequest struct {
	Filename    string `json:"filename" validate:"required,max=255"`
	ContentType string `json:"content_type" validate:"required,max=128"`
	Purpose     string `json:"purpose" validate:"required,oneof=product avatar store brand media"`
}

// PresignUploadResponse is returned to the client for PUT upload to R2.
type PresignUploadResponse struct {
	UploadURL string `json:"upload_url"`
	Key       string `json:"key"`
	PublicURL string `json:"public_url"`
	ExpiresAt string `json:"expires_at"`
	Method    string `json:"method"`
}

// UploadConfigResponse describes whether direct uploads are available.
type UploadConfigResponse struct {
	Enabled     bool `json:"enabled"`
	MaxSizeMB   int  `json:"max_size_mb"`
	PresignTTLs int  `json:"presign_ttl_seconds"`
}
