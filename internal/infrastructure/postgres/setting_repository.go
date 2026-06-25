package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// SettingRepository persists application settings with GORM.
type SettingRepository struct {
	db *gorm.DB
}

// NewSettingRepository creates a GORM-backed settings repository.
func NewSettingRepository(db *gorm.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

// FindByKey loads a setting by key.
func (r *SettingRepository) FindByKey(ctx context.Context, key string) (*models.Setting, error) {
	var setting models.Setting
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// ListAll returns every setting row.
func (r *SettingRepository) ListAll(ctx context.Context) ([]models.Setting, error) {
	var settings []models.Setting
	err := r.db.WithContext(ctx).Find(&settings).Error
	return settings, err
}

// UpdateByKey updates value fields for an existing key.
func (r *SettingRepository) UpdateByKey(ctx context.Context, key string, value json.RawMessage, description *string, updatedAt time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.Setting{}).Where("key = ?", key).Updates(map[string]interface{}{
		"value":       value,
		"description": description,
		"updated_at":  updatedAt,
	})
	return result.RowsAffected, result.Error
}

// Create inserts a new setting row.
func (r *SettingRepository) Create(ctx context.Context, setting *models.Setting) error {
	return r.db.WithContext(ctx).Create(setting).Error
}

// DeleteByKey removes a setting by key.
func (r *SettingRepository) DeleteByKey(ctx context.Context, key string) (int64, error) {
	result := r.db.WithContext(ctx).Where("key = ?", key).Delete(&models.Setting{})
	return result.RowsAffected, result.Error
}
