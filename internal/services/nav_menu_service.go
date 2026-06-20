package services

import (
	"context"
	"encoding/json"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type NavMenuServiceInterface interface {
	GetAll(ctx context.Context) ([]dto.NavItemResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.NavItemResponse, error)
	Create(ctx context.Context, req *dto.UpsertNavMenuRequest) (*dto.NavItemResponse, error)
	Update(ctx context.Context, id uint, req *dto.UpsertNavMenuRequest) (*dto.NavItemResponse, error)
	Reorder(ctx context.Context, req *dto.ReorderNavMenusRequest) error
	Delete(ctx context.Context, id uint) error
}

type navMenuService struct {
	db *gorm.DB
}

func NewNavMenuService(db *gorm.DB) NavMenuServiceInterface {
	return &navMenuService{db: db}
}

// GetAll menu menu items
func (s *navMenuService) GetAll(ctx context.Context) ([]dto.NavItemResponse, error) {
	var menus []models.NavMenu
	if err := s.db.WithContext(ctx).Order("nav_menus.\"order\" ASC").Find(&menus).Error; err != nil {
		return nil, err
	}
	result := make([]dto.NavItemResponse, 0, len(menus))
	for _, m := range menus {
		resp, _ := dto.ToNavItemResponse(&m)
		result = append(result, *resp)
	}
	return result, nil
}

// GetByID get nav menu items with id
func (s *navMenuService) GetByID(ctx context.Context, id uint) (*dto.NavItemResponse, error) {
	var menu models.NavMenu
	if err := s.db.WithContext(ctx).First(&menu, id).Error; err != nil {
		return nil, err
	}
	return dto.ToNavItemResponse(&menu)
}

// Create handle new nav menu items
func (s *navMenuService) Create(ctx context.Context, req *dto.UpsertNavMenuRequest) (*dto.NavItemResponse, error) {
	viewAllJSON, _ := json.Marshal(req.ViewAll)
	columnsJSON, _ := json.Marshal(req.Columns)
	featuredJSON, _ := json.Marshal(req.Featured)

	menu := models.NavMenu{
		Label:    req.Label,
		Type:     req.Type,
		Href:     req.Href,
		Badge:    req.Badge,
		ViewAll:  datatypes.JSON(viewAllJSON),
		Columns:  datatypes.JSON(columnsJSON),
		Featured: datatypes.JSON(featuredJSON),
		SortOrder: req.Order,
	}
	if err := s.db.WithContext(ctx).Create(&menu).Error; err != nil {
		return nil, err
	}
	return dto.ToNavItemResponse(&menu)
}

// Update get nav menu items with id
func (s *navMenuService) Update(ctx context.Context, id uint, req *dto.UpsertNavMenuRequest) (*dto.NavItemResponse, error) {
	var menu models.NavMenu
	if err := s.db.WithContext(ctx).First(&menu, id).Error; err != nil {
		return nil, err
	}
	viewAllJSON, _ := json.Marshal(req.ViewAll)
	columnsJSON, _ := json.Marshal(req.Columns)
	featuredJSON, _ := json.Marshal(req.Featured)

	menu.Label = req.Label
	menu.Type = req.Type
	menu.Href = req.Href
	menu.Badge = req.Badge
	menu.ViewAll = viewAllJSON
	menu.Columns = columnsJSON
	menu.Featured = featuredJSON
	menu.SortOrder = req.Order

	if err := s.db.WithContext(ctx).Save(&menu).Error; err != nil {
		return nil, err
	}
	return dto.ToNavItemResponse(&menu)
}

// Reorder updates display order for multiple nav items in one transaction.
func (s *navMenuService) Reorder(ctx context.Context, req *dto.ReorderNavMenusRequest) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Items {
			result := tx.Model(&models.NavMenu{}).
				Where("id = ?", item.ID).
				Update("order", item.Order)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}
		return nil
	})
}

// Delete removes a nav menu item by ID
func (s *navMenuService) Delete(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&models.NavMenu{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
