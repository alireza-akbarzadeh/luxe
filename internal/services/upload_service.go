package services

import (
	"context"

	appupload "github.com/alireza-akbarzadeh/luxe/internal/application/upload"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
)

type UploadServiceInterface interface {
	IsEnabled() bool
	GetConfig() dto.UploadConfigResponse
	CreatePresignedUpload(ctx context.Context, userID uint, req dto.PresignUploadRequest) (*dto.PresignUploadResponse, error)
}

type uploadService struct {
	inner *appupload.Service
}

func NewUploadService(cfg *config.Config) UploadServiceInterface {
	return &uploadService{inner: appupload.NewService(cfg)}
}

func (s *uploadService) IsEnabled() bool {
	return s.inner.IsEnabled()
}

func (s *uploadService) GetConfig() dto.UploadConfigResponse {
	return s.inner.GetConfig()
}

func (s *uploadService) CreatePresignedUpload(ctx context.Context, userID uint, req dto.PresignUploadRequest) (*dto.PresignUploadResponse, error) {
	return s.inner.CreatePresignedUpload(ctx, userID, req)
}
