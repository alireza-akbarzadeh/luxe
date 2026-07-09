package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// AdminNavRepository persists admin navigation preferences per user.
type AdminNavRepository struct {
	db *gorm.DB
}

// NewAdminNavRepository creates an admin nav preferences repository.
func NewAdminNavRepository(db *gorm.DB) *AdminNavRepository {
	return &AdminNavRepository{db: db}
}

// GetByUserID returns nav preferences for a user, creating defaults if missing.
func (r *AdminNavRepository) GetByUserID(ctx context.Context, userID uint) (*dto.AdminNavPreferencesResponse, error) {
	var row models.AdminNavPreferences
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &dto.AdminNavPreferencesResponse{
			Favorites: []string{},
			Recent:    []dto.AdminNavRecentPage{},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return decodeNavPreferences(&row)
}

// Upsert saves nav preferences for a user.
func (r *AdminNavRepository) Upsert(ctx context.Context, userID uint, req dto.UpdateAdminNavPreferencesRequest) (*dto.AdminNavPreferencesResponse, error) {
	favoritesJSON, err := json.Marshal(req.Favorites)
	if err != nil {
		return nil, err
	}
	recentJSON, err := json.Marshal(req.Recent)
	if err != nil {
		return nil, err
	}

	row := models.AdminNavPreferences{
		UserID:    userID,
		Favorites: favoritesJSON,
		Recent:    recentJSON,
	}

	err = r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Assign(map[string]any{
			"favorites": favoritesJSON,
			"recent":    recentJSON,
		}).
		FirstOrCreate(&row).Error
	if err != nil {
		return nil, err
	}
	return decodeNavPreferences(&row)
}

func decodeNavPreferences(row *models.AdminNavPreferences) (*dto.AdminNavPreferencesResponse, error) {
	resp := &dto.AdminNavPreferencesResponse{
		Favorites: []string{},
		Recent:    []dto.AdminNavRecentPage{},
	}
	if len(row.Favorites) > 0 {
		if err := json.Unmarshal(row.Favorites, &resp.Favorites); err != nil {
			return nil, err
		}
	}
	if len(row.Recent) > 0 {
		if err := json.Unmarshal(row.Recent, &resp.Recent); err != nil {
			return nil, err
		}
	}
	return resp, nil
}
