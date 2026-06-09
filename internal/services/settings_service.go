package services

import (
	"context"
	"errors"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type SettingServiceInterface interface {
	Get(ctx context.Context, key string) (*dto.SettingResponse, error)
	Set(ctx context.Context, key string, req *dto.SetSettingRequest) (*dto.SettingResponse, error)
	List(ctx context.Context) ([]dto.SettingResponse, error)
	Delete(ctx context.Context, key string) error
}

type settingService struct {
	db *gorm.DB
}

func NewSettingService(db *gorm.DB) SettingServiceInterface {
	return &settingService{db: db}
}

func (s *settingService) Get(ctx context.Context, key string) (*dto.SettingResponse, error) {
	var setting models.Setting
	if err := s.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return settingToResponse(&setting), nil
}

func (s *settingService) List(ctx context.Context) ([]dto.SettingResponse, error) {
	var settings []models.Setting
	if err := s.db.WithContext(ctx).Find(&settings).Error; err != nil {
		return nil, err
	}
	resp := make([]dto.SettingResponse, 0, len(settings))
	for _, setting := range settings {
		resp = append(resp, *settingToResponse(&setting))
	}
	return resp, nil
}

// Set creates or updates a setting (upsert).
func (s *settingService) Set(ctx context.Context, key string, req *dto.SetSettingRequest) (*dto.SettingResponse, error) {
	now := time.Now()
	setting := models.Setting{
		Key:         key,
		Value:       req.Value,
		Description: req.Description,
		UpdatedAt:   now,
	}

	// Try update first, if not found then create.
	result := s.db.WithContext(ctx).Model(&models.Setting{}).Where("key = ?", key).Updates(map[string]interface{}{
		"value":       req.Value,
		"description": req.Description,
		"updated_at":  now,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		// Not found, create
		setting.CreatedAt = now
		if err := s.db.WithContext(ctx).Create(&setting).Error; err != nil {
			return nil, err
		}
	}

	// Fetch the fresh record to get all fields (id, timestamps)
	var fresh models.Setting
	if err := s.db.WithContext(ctx).Where("key = ?", key).First(&fresh).Error; err != nil {
		return nil, err
	}
	return settingToResponse(&fresh), nil
}

func (s *settingService) Delete(ctx context.Context, key string) error {
	result := s.db.WithContext(ctx).Where("key = ?", key).Delete(&models.Setting{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func settingToResponse(s *models.Setting) *dto.SettingResponse {
	return &dto.SettingResponse{
		Key:         s.Key,
		Value:       s.Value,
		Description: s.Description,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}
