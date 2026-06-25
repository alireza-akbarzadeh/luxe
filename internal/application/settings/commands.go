package settings

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// Commands orchestrates settings write use cases.
type Commands struct {
	repo *postgres.SettingRepository
}

// NewCommands creates settings command use cases.
func NewCommands(repo *postgres.SettingRepository) *Commands {
	return &Commands{repo: repo}
}

// Set upserts a setting by key.
func (c *Commands) Set(ctx context.Context, key string, req *dto.SetSettingRequest) (*dto.SettingResponse, error) {
	now := time.Now()
	setting := models.Setting{
		Key:         key,
		Value:       req.Value,
		Description: req.Description,
		UpdatedAt:   now,
	}

	rows, err := c.repo.UpdateByKey(ctx, key, req.Value, req.Description, now)
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		setting.CreatedAt = now
		if err := c.repo.Create(ctx, &setting); err != nil {
			return nil, err
		}
	}

	fresh, err := c.repo.FindByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return settingToResponse(fresh), nil
}

// Delete removes a setting by key.
func (c *Commands) Delete(ctx context.Context, key string) error {
	rows, err := c.repo.DeleteByKey(ctx, key)
	if err != nil {
		return err
	}
	if rows == 0 {
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
