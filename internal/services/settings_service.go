package services

import (
	"context"
	"errors"

	appsettings "github.com/alireza-akbarzadeh/luxe/internal/application/settings"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"gorm.io/gorm"
)

type SettingServiceInterface interface {
	Get(ctx context.Context, key string) (*dto.SettingResponse, error)
	Set(ctx context.Context, key string, req *dto.SetSettingRequest) (*dto.SettingResponse, error)
	List(ctx context.Context) ([]dto.SettingResponse, error)
	Delete(ctx context.Context, key string) error
}

type settingService struct {
	commands *appsettings.Commands
	queries  *appsettings.Queries
}

func NewSettingService(db *gorm.DB) SettingServiceInterface {
	repo := postgres.NewSettingRepository(db)
	return &settingService{
		commands: appsettings.NewCommands(repo),
		queries:  appsettings.NewQueries(repo),
	}
}

func (s *settingService) Get(ctx context.Context, key string) (*dto.SettingResponse, error) {
	resp, err := s.queries.Get(ctx, key)
	if errors.Is(err, appsettings.ErrNotFound) {
		return nil, ErrNotFound
	}
	return resp, err
}

func (s *settingService) List(ctx context.Context) ([]dto.SettingResponse, error) {
	return s.queries.List(ctx)
}

func (s *settingService) Set(ctx context.Context, key string, req *dto.SetSettingRequest) (*dto.SettingResponse, error) {
	return s.commands.Set(ctx, key, req)
}

func (s *settingService) Delete(ctx context.Context, key string) error {
	err := s.commands.Delete(ctx, key)
	if errors.Is(err, appsettings.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
