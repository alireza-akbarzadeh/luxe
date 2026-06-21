package services

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

type CompareServiceInterface interface {
	GetCompareList(userID uint) ([]uint, error)
	SyncCompareList(userID uint, productIDs []uint) error
	GetForCompare(ctx context.Context, productIDs []uint) ([]*dto.CompareProductResponse, error)
}

type compareService struct {
	db *gorm.DB
}

func NewCompareService(db *gorm.DB) CompareServiceInterface {
	return &compareService{db: db}
}

func (s *compareService) GetForCompare(ctx context.Context, productIDs []uint) ([]*dto.CompareProductResponse, error) {
	if len(productIDs) < 2 || len(productIDs) > 4 {
		return nil, utils.ErrBadRequest("product count must be between 2 and 4")
	}

	var products []models.Product
	err := s.db.
		Preload("Category").
		Preload("Store").
		Where("id IN ?", productIDs).
		Find(&products).Error
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	if len(products) != len(productIDs) {
		return nil, utils.ErrNotFound("some products not found")
	}

	result := make([]*dto.CompareProductResponse, len(products))
	for i, p := range products {
		discount := 0.0
		if p.CompareAtPrice != nil && *p.CompareAtPrice > p.Price {
			discount = ((*p.CompareAtPrice - p.Price) / *p.CompareAtPrice) * 100
		}

		storeName, storeSlug, storeLogo, shippingInfo, returnPolicy := "", "", "", "", ""
		if p.Store != nil {
			storeName = p.Store.Name
			storeSlug = p.Store.Slug
			storeLogo = p.Store.LogoURL
			shippingInfo = p.Store.ShippingInfo
			returnPolicy = p.Store.ReturnPolicy
		}

		resp := dto.CompareProductResponse{
			ProductResponse: dto.ToProductResponse(ctx, p),
			StoreName:       storeName,
			StoreSlug:       storeSlug,
			StoreLogo:       storeLogo,
			ShippingInfo:    shippingInfo,
			ReturnPolicy:    returnPolicy,
			DiscountPercent: discount,
		}
		result[i] = &resp
	}
	return result, nil
}

func (s *compareService) GetCompareList(userID uint) ([]uint, error) {
	var list models.CompareList
	err := s.db.Where("user_id = ?", userID).First(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []uint{}, nil
	}
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return []uint(list.ProductIDs), nil
}

func (s *compareService) SyncCompareList(userID uint, productIDs []uint) error {
	if len(productIDs) > 4 {
		return utils.ErrBadRequest("cannot compare more than 4 products")
	}
	var list models.CompareList
	err := s.db.Where("user_id = ?", userID).First(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		list = models.CompareList{UserID: userID, ProductIDs: models.UintArray(productIDs)}
		return s.db.Create(&list).Error
	}
	if err != nil {
		return utils.ErrInternal(err)
	}
	list.ProductIDs = models.UintArray(productIDs)
	return s.db.Save(&list).Error
}
