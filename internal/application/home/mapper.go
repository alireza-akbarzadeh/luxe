package home

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// Service serves homepage aggregation queries.
type Service struct {
	storefront *postgres.StorefrontRepository
	home       *postgres.HomeRepository
	cache      *ttlCache
}

// NewService wires homepage use cases.
func NewService(db interface{ /* gorm */ }, storefront *postgres.StorefrontRepository, home *postgres.HomeRepository) *Service {
	_ = db
	return &Service{
		storefront: storefront,
		home:       home,
		cache:      newTTLCache(),
	}
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > 24 {
		return 24
	}
	return limit
}

func discountPercent(price float64, compare *float64) *int {
	if compare == nil || *compare <= price {
		return nil
	}
	pct := int(math.Round(((*compare - price) / *compare) * 100))
	return &pct
}

func toHomeProduct(ctx context.Context, p *models.Product, unitsSold int64, wishlist int64) dto.HomeProductItem {
	if p == nil {
		return dto.HomeProductItem{}
	}
	base := dto.ToProductResponse(ctx, *p)
	item := dto.HomeProductItem{
		ProductResponse: base,
		UnitsSold:       unitsSold,
		WishlistCount:   wishlist,
		DiscountPercent: discountPercent(p.Price, p.CompareAtPrice),
	}
	return item
}

func productsToHome(ctx context.Context, products []*models.Product) []dto.HomeProductItem {
	out := make([]dto.HomeProductItem, 0, len(products))
	for _, p := range products {
		if p == nil {
			continue
		}
		out = append(out, toHomeProduct(ctx, p, 0, 0))
	}
	return out
}

func railTitleForCategory(name string) string {
	n := strings.TrimSpace(name)
	if n == "" {
		return "Recommended for You"
	}
	return fmt.Sprintf("Best in %s", n)
}

func trendingTitleForCategory(name string) string {
	n := strings.TrimSpace(name)
	if n == "" {
		return "Trending Now"
	}
	return fmt.Sprintf("Trending %s", n)
}
