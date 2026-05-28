package services

import (
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

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

// GlobalSearch performs full‑text search on products, stores, and categories concurrently.
func (s *searchService) GlobalSearch(req dto.SearchRequest) (*dto.SearchResponse, error) {
	if req.Query == "" {
		return &dto.SearchResponse{}, nil
	}

	type result struct {
		products   []*models.Product
		stores     []*models.Store
		categories []*models.Category
		err        error
	}
	ch := make(chan result, 3)

	// Search products with filters
	go func() {
		query := s.db.
			Where("search_vector @@ plainto_tsquery('english', ?)", req.Query).
			Order(gorm.Expr("ts_rank(search_vector, plainto_tsquery('english', ?)) DESC", req.Query))

		// Apply product filters

		if req.CategorySlug != "" {
			query = query.Joins("JOIN categories ON categories.id = products.category_id").
				Where("categories.slug = ?", req.CategorySlug)
		}

		if req.CategoryID != nil {
			query = query.Where("category_id = ?", *req.CategoryID)
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
		switch req.Sort {
		case "price_asc":
			query = query.Order("price ASC")
		case "price_desc":
			query = query.Order("price DESC")
		case "rating_desc":
			query = query.Order("rating DESC")
		case "newest":
			query = query.Order("created_at DESC")
		}

		var products []*models.Product
		err := query.Limit(req.Limit).Offset(req.Offset).
			Preload("Category").
			Find(&products).Error
		ch <- result{products: products, err: err}
	}()

	// Search stores (only by search query, no additional filters)
	go func() {
		var stores []*models.Store
		err := s.db.
			Where("search_vector @@ plainto_tsquery('english', ?)", req.Query).
			Order(gorm.Expr("ts_rank(search_vector, plainto_tsquery('english', ?)) DESC", req.Query)).
			Limit(req.Limit).Offset(req.Offset).
			Preload("Categories").
			Find(&stores).Error
		ch <- result{stores: stores, err: err}
	}()

	// Search categories (only by search query, no additional filters)
	go func() {
		var categories []*models.Category
		err := s.db.
			Where("search_vector @@ plainto_tsquery('english', ?)", req.Query).
			Order(gorm.Expr("ts_rank(search_vector, plainto_tsquery('english', ?)) DESC", req.Query)).
			Limit(req.Limit).Offset(req.Offset).
			Find(&categories).Error
		ch <- result{categories: categories, err: err}
	}()

	var products []*models.Product
	var stores []*models.Store
	var categories []*models.Category
	for i := 0; i < 3; i++ {
		res := <-ch
		if res.err != nil {
			return nil, utils.ErrInternal(res.err)
		}
		products = append(products, res.products...)
		stores = append(stores, res.stores...)
		categories = append(categories, res.categories...)
	}
	close(ch)

	total := int64(len(products) + len(stores) + len(categories))

	return &dto.SearchResponse{
		Products:   dto.ToProductResponses(products),
		Stores:     dto.ToStoreResponses(stores),
		Categories: dto.ToCategoryResponses(categories),
		Total:      total,
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
	s.db.Where("name ILIKE ?", "%"+query+"%").
		Limit(limitPerType).
		Find(&products)
	for _, p := range products {
		img := p.Images[0]
		suggestions = append(suggestions, dto.SuggestionItem{
			Type:  "product",
			ID:    &p.ID,
			Name:  p.Name,
			Slug:  p.Slug,
			Image: &img,
			Price: &p.Price,
		})
	}

	// Store suggestions
	var stores []models.Store
	s.db.Where("name ILIKE ?", "%"+query+"%").
		Limit(limitPerType).
		Find(&stores)
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
	s.db.Where("name ILIKE ?", "%"+query+"%").
		Limit(limitPerType).
		Find(&categories)
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
	log := models.SearchLog{
		Query:  query,
		UserID: userID,
	}
	return s.db.Create(&log).Error
}
