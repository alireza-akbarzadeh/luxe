package settings

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
)

// ErrNotFound is returned when a setting key does not exist.
var ErrNotFound = errors.New("resource not found")

// Queries orchestrates settings read use cases.
type Queries struct {
	repo *postgres.SettingRepository
}

// NewQueries creates settings query use cases.
func NewQueries(repo *postgres.SettingRepository) *Queries {
	return &Queries{repo: repo}
}

// Get returns a single setting by key.
func (q *Queries) Get(ctx context.Context, key string) (*dto.SettingResponse, error) {
	setting, err := q.repo.FindByKey(ctx, key)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return settingToResponse(setting), nil
}

// List returns all settings.
func (q *Queries) List(ctx context.Context) ([]dto.SettingResponse, error) {
	settings, err := q.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.SettingResponse, 0, len(settings))
	for i := range settings {
		resp = append(resp, *settingToResponse(&settings[i]))
	}
	return resp, nil
}
