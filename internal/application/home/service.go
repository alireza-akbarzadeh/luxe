package home

import (
	"context"
	"encoding/json"
	"fmt"

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
func NewService(storefront *postgres.StorefrontRepository, home *postgres.HomeRepository) *Service {
	return &Service{
		storefront: storefront,
		home:       home,
		cache:      newTTLCache(),
	}
}

// Manifest returns lightweight section keys for clients.
func (s *Service) Manifest() dto.HomeManifestResponse {
	return dto.HomeManifestResponse{
		Sections: []string{
			"categories", "top-brands", "top-products", "trending-products",
			"new-arrivals", "flash-deals", "recommended", "recently-viewed",
			"most-wishlisted", "customer-favorites", "popular-collections",
			"featured-stores", "shop-by-price", "seasonal-picks", "recently-restocked",
		},
		CacheTTL: int(publicCacheTTL.Seconds()),
	}
}

// GetCategories returns popular, featured, and optional personalized category data.
func (s *Service) GetCategories(ctx context.Context, userID *uint, limit int) (dto.HomeCategoriesResponse, error) {
	limit = clampLimit(limit)
	cacheKey := fmt.Sprintf("home:categories:%d", limit)
	if userID == nil {
		if cached, ok := s.cache.Get(cacheKey); ok {
			return cached.(dto.HomeCategoriesResponse), nil
		}
	}

	rows, err := s.storefront.ListHomeCategories(ctx, limit)
	if err != nil {
		return dto.HomeCategoriesResponse{}, err
	}

	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	covers, _ := s.storefront.FindCategoryCoverImages(ctx, ids)

	popular := make([]dto.HomeCategoryItem, 0, len(rows))
	for _, row := range rows {
		popular = append(popular, dto.HomeCategoryItem{
			ID: row.ID, Name: row.Name, Slug: row.Slug, Description: row.Description,
			ImageURL: covers[row.ID], ProductCount: row.ProductCount,
		})
	}

	featuredCount := 4
	if featuredCount > len(popular) {
		featuredCount = len(popular)
	}
	featured := append([]dto.HomeCategoryItem(nil), popular[:featuredCount]...)

	resp := dto.HomeCategoriesResponse{
		Popular:  popular,
		Featured: featured,
	}

	if userID != nil {
		favs, err := s.home.ListFavoriteCategories(ctx, *userID)
		if err != nil {
			return resp, err
		}
		favIDs := make([]uint, 0, len(favs))
		favItems := make([]dto.HomeCategoryItem, 0, len(favs))
		for _, c := range favs {
			favIDs = append(favIDs, c.ID)
			img := covers[c.ID]
			favItems = append(favItems, dto.HomeCategoryItem{
				ID: c.ID, Name: c.Name, Slug: c.Slug, Description: c.Description, ImageURL: img,
			})
		}
		resp.Favorite = favItems
		if len(favIDs) > 0 {
			favCovers, _ := s.storefront.FindCategoryCoverImages(ctx, favIDs)
			for i := range resp.Favorite {
				if resp.Favorite[i].ImageURL == "" {
					resp.Favorite[i].ImageURL = favCovers[resp.Favorite[i].ID]
				}
			}
			resp.PersonalizedRails = s.buildPersonalizedRails(ctx, favs, 8)
		}
	} else {
		s.cache.Set(cacheKey, resp, publicCacheTTL)
	}

	return resp, nil
}

func (s *Service) buildPersonalizedRails(ctx context.Context, categories []models.Category, perRail int) []dto.HomeProductRail {
	rails := make([]dto.HomeProductRail, 0, len(categories))
	for _, cat := range categories {
		products, err := s.storefront.ListProductsByCategory(ctx, cat.ID, perRail)
		if err != nil || len(products) == 0 {
			continue
		}
		rails = append(rails, dto.HomeProductRail{
			Key:      cat.Slug,
			Title:    railTitleForCategory(cat.Name),
			Products: productsToHome(ctx, products),
		})
	}
	return rails
}

// GetTopBrands returns sales-ranked brands.
func (s *Service) GetTopBrands(ctx context.Context, limit int) ([]dto.HomeBrandItem, error) {
	limit = clampLimit(limit)
	cacheKey := fmt.Sprintf("home:top-brands:%d", limit)
	if cached, ok := s.cache.Get(cacheKey); ok {
		return cached.([]dto.HomeBrandItem), nil
	}

	rows, err := s.storefront.ListTopBrandsBySales(ctx, limit)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	banners, _ := s.storefront.FindBrandBannerImages(ctx, ids)

	out := make([]dto.HomeBrandItem, 0, len(rows))
	for _, row := range rows {
		logo := row.LogoURL
		item := dto.HomeBrandItem{
			BrandResponse: dto.BrandResponse{
				ID: row.ID, Name: row.Name, Slug: row.Slug, Description: row.Description,
				LogoURL: logo, Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			},
			BannerURL:    banners[row.ID],
			ProductCount: row.ProductCount,
			UnitsSold:    row.UnitsSold,
			Revenue:      row.Revenue,
			MinPrice:     row.MinPrice,
		}
		out = append(out, item)
	}
	s.cache.Set(cacheKey, out, publicCacheTTL)
	return out, nil
}

// GetTopProducts returns best-selling products.
func (s *Service) GetTopProducts(ctx context.Context, limit int) ([]dto.HomeProductItem, error) {
	limit = clampLimit(limit)
	cacheKey := fmt.Sprintf("home:top-products:%d", limit)
	if cached, ok := s.cache.Get(cacheKey); ok {
		return cached.([]dto.HomeProductItem), nil
	}

	rows, err := s.storefront.ListBestSellerIDs(ctx, limit)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(rows))
	soldMap := make(map[uint]int64, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ProductID)
		soldMap[row.ProductID] = row.UnitsSold
	}
	products, err := s.storefront.ListProductsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]dto.HomeProductItem, 0, len(products))
	for _, p := range products {
		out = append(out, toHomeProduct(ctx, p, soldMap[p.ID], 0))
	}
	s.cache.Set(cacheKey, out, publicCacheTTL)
	return out, nil
}

// GetTrendingProducts returns trending products.
func (s *Service) GetTrendingProducts(ctx context.Context, limit int) ([]dto.HomeProductItem, error) {
	limit = clampLimit(limit)
	cacheKey := fmt.Sprintf("home:trending:%d", limit)
	if cached, ok := s.cache.Get(cacheKey); ok {
		return cached.([]dto.HomeProductItem), nil
	}

	rows, err := s.storefront.ListTrendingProductIDs(ctx, limit)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ProductID)
	}
	products, err := s.storefront.ListProductsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := productsToHome(ctx, products)
	s.cache.Set(cacheKey, out, publicCacheTTL)
	return out, nil
}

// GetNewArrivals returns newest products.
func (s *Service) GetNewArrivals(ctx context.Context, limit int) ([]dto.HomeProductItem, error) {
	limit = clampLimit(limit)
	cacheKey := fmt.Sprintf("home:new-arrivals:%d", limit)
	if cached, ok := s.cache.Get(cacheKey); ok {
		return cached.([]dto.HomeProductItem), nil
	}
	products, err := s.storefront.ListNewArrivalProducts(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := productsToHome(ctx, products)
	s.cache.Set(cacheKey, out, publicCacheTTL)
	return out, nil
}

// GetFlashDeals returns active flash deals.
func (s *Service) GetFlashDeals(ctx context.Context, limit int) ([]dto.HomeFlashDealItem, error) {
	limit = clampLimit(limit)
	deals, err := s.home.ListActiveFlashDeals(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]dto.HomeFlashDealItem, 0, len(deals))
	for _, d := range deals {
		if d.Product == nil {
			continue
		}
		ends := d.EndsAt
		item := dto.HomeFlashDealItem{
			ID:            d.ID,
			EndsAt:        ends,
			QuantityLimit: d.QuantityLimit,
			Product:       toHomeProduct(ctx, d.Product, 0, 0),
		}
		item.Product.FlashEndsAt = &ends
		item.Product.QuantityLimit = d.QuantityLimit
		out = append(out, item)
	}
	return out, nil
}

// GetMostWishlisted returns products ranked by wishlist count.
func (s *Service) GetMostWishlisted(ctx context.Context, limit int) ([]dto.HomeProductItem, error) {
	limit = clampLimit(limit)
	rows, err := s.storefront.ListMostWishlistedIDs(ctx, limit)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(rows))
	wishMap := make(map[uint]int64)
	for _, row := range rows {
		ids = append(ids, row.ProductID)
		wishMap[row.ProductID] = row.LikeCount
	}
	products, err := s.storefront.ListProductsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]dto.HomeProductItem, 0, len(products))
	for _, p := range products {
		out = append(out, toHomeProduct(ctx, p, 0, wishMap[p.ID]))
	}
	return out, nil
}

// GetCustomerFavorites returns highest-rated products.
func (s *Service) GetCustomerFavorites(ctx context.Context, limit int) ([]dto.HomeProductItem, error) {
	limit = clampLimit(limit)
	products, err := s.storefront.ListCustomerFavoriteProducts(ctx, limit)
	if err != nil {
		return nil, err
	}
	return productsToHome(ctx, products), nil
}

// GetPopularCollections returns published collections.
func (s *Service) GetPopularCollections(ctx context.Context, limit int) ([]dto.HomeCollectionItem, error) {
	limit = clampLimit(limit)
	rows, err := s.home.ListPublishedCollections(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]dto.HomeCollectionItem, 0, len(rows))
	for _, c := range rows {
		out = append(out, dto.HomeCollectionItem{
			ID: c.ID, Slug: c.Slug, Eyebrow: c.Eyebrow, Title: c.Title,
			Description: c.Description, Href: c.Href, ImageURL: c.ImageURL, CTALabel: c.CTALabel,
		})
	}
	return out, nil
}

// GetFeaturedStores returns top-performing stores.
func (s *Service) GetFeaturedStores(ctx context.Context, limit int) ([]dto.HomeStoreItem, error) {
	limit = clampLimit(limit)
	rows, err := s.storefront.ListFeaturedStores(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]dto.HomeStoreItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.HomeStoreItem{
			ID: row.ID, Name: row.Name, Slug: row.Slug, LogoURL: row.LogoURL,
			BannerURL: row.BannerURL, Rating: row.Rating, ReviewCount: row.ReviewCount,
			ProductCount: row.ProductCount, UnitsSold: row.UnitsSold, Revenue: row.Revenue,
		})
	}
	return out, nil
}

// GetShopByPrice returns static price bucket cards.
func (s *Service) GetShopByPrice(ctx context.Context) ([]dto.HomeShopByPriceItem, error) {
	buckets := []struct {
		key, title string
		max        float64
	}{
		{"under-25", "Under $25", 25},
		{"under-50", "Under $50", 50},
		{"under-100", "Under $100", 100},
		{"luxury", "Luxury Picks", 999999},
	}
	out := make([]dto.HomeShopByPriceItem, 0, len(buckets))
	for _, b := range buckets {
		href := fmt.Sprintf("/shop?maxPrice=%.0f", b.max)
		if b.key == "luxury" {
			href = "/shop?sort=rating_desc"
		}
		count := int64(0)
		if b.key != "luxury" {
			products, _ := s.storefront.ListProductsByMaxPrice(ctx, b.max, 1)
			if len(products) > 0 {
				count = 1
			}
		}
		out = append(out, dto.HomeShopByPriceItem{
			Key: b.key, Title: b.title, MaxPrice: b.max, Href: href, PreviewCount: count,
		})
	}
	return out, nil
}

// GetSeasonalPicks returns published homepage sections.
func (s *Service) GetSeasonalPicks(ctx context.Context, limit int) ([]dto.HomeSectionItem, error) {
	limit = clampLimit(limit)
	rows, err := s.home.ListPublishedHomepageSections(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]dto.HomeSectionItem, 0, len(rows))
	for _, row := range rows {
		filters := map[string]interface{}{}
		if len(row.Filters) > 0 {
			_ = json.Unmarshal(row.Filters, &filters)
		}
		out = append(out, dto.HomeSectionItem{
			Key: row.SectionKey, Title: row.Title, Href: row.Href, ImageURL: row.ImageURL, Filters: filters,
		})
	}
	return out, nil
}

// GetRecentlyRestocked returns recently restocked products.
func (s *Service) GetRecentlyRestocked(ctx context.Context, limit int) ([]dto.HomeProductItem, error) {
	limit = clampLimit(limit)
	ids, err := s.storefront.ListRecentlyRestockedProductIDs(ctx, limit)
	if err != nil {
		return nil, err
	}
	products, err := s.storefront.ListProductsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return productsToHome(ctx, products), nil
}

// GetRecentlyViewed returns products the user recently viewed.
func (s *Service) GetRecentlyViewed(ctx context.Context, userID uint, limit int) ([]dto.HomeProductItem, error) {
	limit = clampLimit(limit)
	ids, err := s.home.ListRecentlyViewedProductIDs(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	products, err := s.storefront.ListProductsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return productsToHome(ctx, products), nil
}

// RecordProductView upserts a product view for personalization.
func (s *Service) RecordProductView(ctx context.Context, userID, productID uint) error {
	return s.home.UpsertProductView(ctx, userID, productID)
}

// GetRecommended returns personalized product recommendations.
func (s *Service) GetRecommended(ctx context.Context, userID uint, limit int) ([]dto.HomeProductItem, error) {
	limit = clampLimit(limit)
	cacheKey := fmt.Sprintf("home:recommended:%d:%d", userID, limit)
	if cached, ok := s.cache.Get(cacheKey); ok {
		return cached.([]dto.HomeProductItem), nil
	}

	seen := make(map[uint]struct{})
	ids := make([]uint, 0, limit)

	appendIDs := func(list []uint) {
		for _, id := range list {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
			if len(ids) >= limit {
				return
			}
		}
	}

	viewed, _ := s.home.ListRecentlyViewedProductIDs(ctx, userID, limit)
	appendIDs(viewed)
	liked, _ := s.home.ListUserLikedProductIDs(ctx, userID, limit)
	appendIDs(liked)
	catIDs, _ := s.home.ListUserOrderCategoryIDs(ctx, userID, 3)
	for _, catID := range catIDs {
		prods, _ := s.storefront.ListProductsByCategory(ctx, catID, 4)
		for _, p := range prods {
			appendIDs([]uint{p.ID})
		}
	}
	favCats, _ := s.home.ListFavoriteCategoryIDs(ctx, userID)
	for _, catID := range favCats {
		prods, _ := s.storefront.ListProductsByCategory(ctx, catID, 4)
		for _, p := range prods {
			appendIDs([]uint{p.ID})
		}
	}
	if len(ids) < limit {
		trending, _ := s.storefront.ListTrendingProductIDs(ctx, limit)
		for _, row := range trending {
			appendIDs([]uint{row.ProductID})
		}
	}

	products, err := s.storefront.ListProductsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := productsToHome(ctx, products)
	s.cache.Set(cacheKey, out, personalCacheTTL)
	return out, nil
}

// GetFavoriteCategories returns user's saved favorite categories.
func (s *Service) GetFavoriteCategories(ctx context.Context, userID uint) ([]dto.HomeCategoryItem, error) {
	categories, err := s.home.ListFavoriteCategories(ctx, userID)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(categories))
	for _, c := range categories {
		ids = append(ids, c.ID)
	}
	covers, _ := s.storefront.FindCategoryCoverImages(ctx, ids)
	out := make([]dto.HomeCategoryItem, 0, len(categories))
	for _, c := range categories {
		out = append(out, dto.HomeCategoryItem{
			ID: c.ID, Name: c.Name, Slug: c.Slug, Description: c.Description, ImageURL: covers[c.ID],
		})
	}
	return out, nil
}

// SetFavoriteCategories replaces favorite categories for a user.
func (s *Service) SetFavoriteCategories(ctx context.Context, userID uint, categoryIDs []uint) error {
	return s.home.ReplaceFavoriteCategories(ctx, userID, categoryIDs)
}
