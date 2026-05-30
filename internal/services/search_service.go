package services

import (
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

const searchFacetLimit = 10

type SearchServiceInterface interface {
	GlobalSearch(req dto.SearchRequest) (*dto.SearchResponse, error)
	Suggestions(query string, limit int) (*dto.SuggestionsResponse, error)
	Trending(limit int) ([]dto.TrendingSearch, error)
	LogSearch(query string, userID *uint) error
}

type searchService struct {
	db *gorm.DB
}

func NewSearchService(db *gorm.DB) SearchServiceInterface {
	return &searchService{db: db}
}

func (s *searchService) applyProductFilters(query *gorm.DB, req dto.SearchRequest) *gorm.DB {
	if req.Query != "" {
		query = query.Where("search_vector @@ plainto_tsquery('english', ?)", req.Query)
	} else {
		query = query.Where("status = ?", "active")
	}

	if req.CategorySlug != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.slug = ?", req.CategorySlug)
	}
	if req.CategoryID != nil {
		query = query.Where("category_id = ?", *req.CategoryID)
	}
	if req.StoreID != nil {
		query = query.Where("store_id = ?", *req.StoreID)
	}
	if req.MinPrice > 0 {
		query = query.Where("price >= ?", req.MinPrice)
	}
	if req.MaxPrice > 0 {
		query = query.Where("price <= ?", req.MaxPrice)
	}
	if req.MinRating > 0 {
		query = query.Where("rating >= ?", req.MinRating)
	}
	if req.IsDigital != nil {
		query = query.Where("is_digital = ?", *req.IsDigital)
	}
	if req.IsNew != nil {
		query = query.Where("is_new = ?", *req.IsNew)
	}
	if req.InStock != nil && *req.InStock {
		query = query.Where("stock > 0")
	}
	if req.OnSale != nil && *req.OnSale {
		query = query.Where("compare_at_price IS NOT NULL AND compare_at_price > price")
	}

	return query
}

func (s *searchService) applyProductSort(query *gorm.DB, req dto.SearchRequest) *gorm.DB {
	switch req.Sort {
	case "price_asc":
		return query.Order("price ASC")
	case "price_desc":
		return query.Order("price DESC")
	case "rating_desc":
		return query.Order("rating DESC")
	case "newest":
		return query.Order("created_at DESC")
	case "popular":
		return query.Order("reviews_count DESC, rating DESC")
	default:
		if req.Query != "" {
			return query.Order(gorm.Expr("ts_rank(search_vector, plainto_tsquery('english', ?)) DESC", req.Query))
		}
		return query.Order("created_at DESC")
	}
}

// GlobalSearch performs full-text search on products, stores, and categories concurrently.
func (s *searchService) GlobalSearch(req dto.SearchRequest) (*dto.SearchResponse, error) {
	type result struct {
		products   []*models.Product
		productTotal int64
		stores     []*models.Store
		categories []*models.Category
		err        error
	}
	ch := make(chan result, 3)

	go func() {
		countQuery := s.applyProductFilters(s.db.Model(&models.Product{}), req)

		var productTotal int64
		if err := countQuery.Count(&productTotal).Error; err != nil {
			ch <- result{err: err}
			return
		}

		findQuery := s.applyProductFilters(s.db.Model(&models.Product{}), req)
		var products []*models.Product
		err := s.applyProductSort(findQuery, req).
			Limit(req.Limit).
			Offset(req.Offset).
			Preload("Category").
			Find(&products).Error
		ch <- result{products: products, productTotal: productTotal, err: err}
	}()

	go func() {
		if req.Query == "" {
			ch <- result{}
			return
		}

		var stores []*models.Store
		err := s.db.
			Where("search_vector @@ plainto_tsquery('english', ?)", req.Query).
			Order(gorm.Expr("ts_rank(search_vector, plainto_tsquery('english', ?)) DESC", req.Query)).
			Limit(searchFacetLimit).
			Preload("Categories").
			Find(&stores).Error
		ch <- result{stores: stores, err: err}
	}()

	go func() {
		if req.Query == "" {
			ch <- result{}
			return
		}

		var categories []*models.Category
		err := s.db.
			Where("search_vector @@ plainto_tsquery('english', ?)", req.Query).
			Order(gorm.Expr("ts_rank(search_vector, plainto_tsquery('english', ?)) DESC", req.Query)).
			Limit(searchFacetLimit).
			Find(&categories).Error
		ch <- result{categories: categories, err: err}
	}()

	var products []*models.Product
	var productTotal int64
	var stores []*models.Store
	var categories []*models.Category
	for i := 0; i < 3; i++ {
		res := <-ch
		if res.err != nil {
			return nil, utils.ErrInternal(res.err)
		}
		if res.products != nil {
			products = res.products
			productTotal = res.productTotal
		}
		stores = append(stores, res.stores...)
		categories = append(categories, res.categories...)
	}

	return &dto.SearchResponse{
		Products:   dto.ToProductResponses(products),
		Stores:     dto.ToStoreResponses(stores),
		Categories: dto.ToCategoryResponses(categories),
		Total:      productTotal,
	}, nil
}

// Suggestions returns a small list of matches for autocomplete (uses faster ILIKE with trigram).
func (s *searchService) Suggestions(query string, limit int) (*dto.SuggestionsResponse, error) {
	if query == "" {
		return &dto.SuggestionsResponse{}, nil
	}
	limitPerType := limit/3 + 1 // ensure we have enough

	var suggestions []dto.SuggestionItem

	// Product suggestions (name only, ILIKE with trigram index)
	var products []models.Product
	if err := s.db.Where("name ILIKE ?", "%"+query+"%").
		Limit(limitPerType).
		Find(&products).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	for _, p := range products {
		var img *string
		if len(p.Images) > 0 {
			img = &p.Images[0]
		}
		suggestions = append(suggestions, dto.SuggestionItem{
			Type:  "product",
			ID:    &p.ID,
			Name:  p.Name,
			Slug:  p.Slug,
			Image: img,
			Price: &p.Price,
		})
	}

	// Store suggestions
	var stores []models.Store
	if err := s.db.Where("name ILIKE ?", "%"+query+"%").
		Limit(limitPerType).
		Find(&stores).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	for _, st := range stores {
		suggestions = append(suggestions, dto.SuggestionItem{
			Type:  "store",
			ID:    &st.ID,
			Name:  st.Name,
			Slug:  st.Slug,
			Image: &st.LogoURL,
		})
	}

	// Category suggestions
	var categories []models.Category
	if err := s.db.Where("name ILIKE ?", "%"+query+"%").
		Limit(limitPerType).
		Find(&categories).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	for _, cat := range categories {
		suggestions = append(suggestions, dto.SuggestionItem{
			Type: "category",
			ID:   &cat.ID,
			Name: cat.Name,
			Slug: cat.Slug,
		})
	}

	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}
	return &dto.SuggestionsResponse{Suggestions: suggestions}, nil
}

// Trending returns the most searched queries from the last 7 days.
func (s *searchService) Trending(limit int) ([]dto.TrendingSearch, error) {
	var results []dto.TrendingSearch
	err := s.db.Model(&models.SearchLog{}).
		Select("query, COUNT(*) as count").
		Where("created_at > NOW() - INTERVAL '7 days'").
		Group("query").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return results, nil
}

// LogSearch records a search query (for trending).
func (s *searchService) LogSearch(query string, userID *uint) error {
	if query == "" {
		return nil
	}
	log := models.SearchLog{
		Query:  query,
		UserID: userID,
	}
	return s.db.Create(&log).Error
}
