package services

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	r2integration "github.com/alireza-akbarzadeh/luxe/internal/integrations/r2"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/google/uuid"
)

var allowedUploadContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

var safeFilenamePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

type UploadServiceInterface interface {
	IsEnabled() bool
	GetConfig() dto.UploadConfigResponse
	CreatePresignedUpload(ctx context.Context, userID uint, req dto.PresignUploadRequest) (*dto.PresignUploadResponse, error)
}

type uploadService struct {
	cfg    *config.Config
	client *r2integration.Client
}

func NewUploadService(cfg *config.Config) UploadServiceInterface {
	if cfg == nil || !cfg.R2.Enabled {
		return &uploadService{cfg: cfg}
	}

	client, err := r2integration.NewClient(&cfg.R2)
	if err != nil {
		utils.Log.WithError(err).Error("failed to initialize R2 client; uploads disabled")
		return &uploadService{cfg: cfg}
	}

	return &uploadService{cfg: cfg, client: client}
}

func (s *uploadService) IsEnabled() bool {
	return s.client != nil
}

func (s *uploadService) GetConfig() dto.UploadConfigResponse {
	if s.cfg == nil {
		return dto.UploadConfigResponse{}
	}
	ttl := int(s.cfg.R2.PresignTTL.Seconds())
	if ttl <= 0 {
		ttl = 900
	}
	return dto.UploadConfigResponse{
		Enabled:     s.IsEnabled(),
		MaxSizeMB:   s.cfg.R2.MaxUploadMB,
		PresignTTLs: ttl,
	}
}

func (s *uploadService) CreatePresignedUpload(ctx context.Context, userID uint, req dto.PresignUploadRequest) (*dto.PresignUploadResponse, error) {
	if !s.IsEnabled() {
		return nil, utils.ErrBadRequest("direct uploads are not configured")
	}

	contentType := strings.TrimSpace(strings.ToLower(req.ContentType))
	if !allowedUploadContentTypes[contentType] {
		return nil, utils.ErrBadRequest("unsupported content type; allowed: jpeg, png, webp, gif")
	}

	filename := sanitizeUploadFilename(req.Filename)
	if filename == "" {
		return nil, utils.ErrBadRequest("invalid filename")
	}

	purpose := strings.TrimSpace(req.Purpose)
	key := fmt.Sprintf("uploads/%s/%d/%s-%s", purpose, userID, uuid.New().String(), filename)

	result, err := s.client.PresignPutObject(ctx, key, contentType)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	return &dto.PresignUploadResponse{
		UploadURL: result.UploadURL,
		Key:       result.Key,
		PublicURL: result.PublicURL,
		ExpiresAt: result.ExpiresAt.Format(time.RFC3339),
		Method:    "PUT",
	}, nil
}

func sanitizeUploadFilename(name string) string {
	base := path.Base(strings.TrimSpace(name))
	base = safeFilenamePattern.ReplaceAllString(base, "_")
	base = strings.Trim(base, "._")
	if base == "" || base == "." {
		return ""
	}
	if len(base) > 120 {
		ext := path.Ext(base)
		base = base[:120-len(ext)] + ext
	}
	return base
}
