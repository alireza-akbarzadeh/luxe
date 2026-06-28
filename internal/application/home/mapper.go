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

func capCategoryItems(items []dto.HomeCategoryItem, limit int) []dto.HomeCategoryItem {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func mergeUniqueCategoryIDs(lists ...[]uint) []uint {
	seen := make(map[uint]struct{})
	out := make([]uint, 0)
	for _, list := range lists {
		for _, id := range list {
			if id == 0 {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

func categoriesToHomeItems(ctx context.Context, storefront *postgres.StorefrontRepository, categories []models.Category) []dto.HomeCategoryItem {
	if len(categories) == 0 {
		return []dto.HomeCategoryItem{}
	}
	ids := make([]uint, 0, len(categories))
	for _, c := range categories {
		ids = append(ids, c.ID)
	}
	covers, _ := storefront.FindCategoryCoverImages(ctx, ids)
	out := make([]dto.HomeCategoryItem, 0, len(categories))
	for _, c := range categories {
		out = append(out, dto.HomeCategoryItem{
			ID: c.ID, Name: c.Name, Slug: c.Slug, Description: c.Description,
			ImageURL: covers[c.ID],
		})
	}
	return out
}
