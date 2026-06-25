package compare

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Queries orchestrates compare read use cases.
type Queries struct {
	repo *postgres.CompareRepository
}

// NewQueries creates compare query use cases.
func NewQueries(repo *postgres.CompareRepository) *Queries {
	return &Queries{repo: repo}
}

// GetCompareList returns product ids in the user's compare list.
func (q *Queries) GetCompareList(userID uint) ([]uint, error) {
	list, err := q.repo.FindCompareListByUser(userID)
	if postgres.IsNotFound(err) {
		return []uint{}, nil
	}
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return []uint(list.ProductIDs), nil
}

// GetForCompare loads product details for side-by-side comparison.
func (q *Queries) GetForCompare(ctx context.Context, productIDs []uint) ([]*dto.CompareProductResponse, error) {
	if len(productIDs) < 2 || len(productIDs) > 4 {
		return nil, utils.ErrBadRequest("product count must be between 2 and 4")
	}

	products, err := q.repo.FindProductsByIDs(ctx, productIDs)
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

// Commands orchestrates compare write use cases.
type Commands struct {
	repo *postgres.CompareRepository
}

// NewCommands creates compare command use cases.
func NewCommands(repo *postgres.CompareRepository) *Commands {
	return &Commands{repo: repo}
}

// SyncCompareList upserts the user's compare list product ids.
func (c *Commands) SyncCompareList(userID uint, productIDs []uint) error {
	if len(productIDs) > 4 {
		return utils.ErrBadRequest("cannot compare more than 4 products")
	}

	list, err := c.repo.FindCompareListByUser(userID)
	if postgres.IsNotFound(err) {
		newList := models.CompareList{UserID: userID, ProductIDs: models.UintArray(productIDs)}
		return c.repo.CreateCompareList(&newList)
	}
	if err != nil {
		return utils.ErrInternal(err)
	}

	list.ProductIDs = models.UintArray(productIDs)
	return c.repo.SaveCompareList(list)
}
