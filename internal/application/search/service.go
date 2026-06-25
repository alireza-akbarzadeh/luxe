package search

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Queries orchestrates search read use cases.
type Queries struct {
	repo *postgres.SearchRepository
}

// NewQueries creates search query use cases.
func NewQueries(repo *postgres.SearchRepository) *Queries {
	return &Queries{repo: repo}
}

// GlobalSearch performs full-text search on products, stores, and categories concurrently.
func (q *Queries) GlobalSearch(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
	type result struct {
		products     []*models.Product
		productTotal int64
		stores       []*models.Store
		categories   []*models.Category
		err          error
	}
	ch := make(chan result, 3)

	go func() {
		total, err := q.repo.CountProducts(req)
		if err != nil {
			ch <- result{err: err}
			return
		}
		products, err := q.repo.FindProducts(req)
		ch <- result{products: products, productTotal: total, err: err}
	}()

	go func() {
		if req.Query == "" {
			ch <- result{}
			return
		}
		stores, err := q.repo.FindStoresByQuery(req.Query)
		ch <- result{stores: stores, err: err}
	}()

	go func() {
		if req.Query == "" {
			ch <- result{}
			return
		}
		categories, err := q.repo.FindCategoriesByQuery(req.Query)
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
		Products:   dto.ToProductResponses(ctx, products),
		Stores:     dto.ToStoreResponses(ctx, stores),
		Categories: dto.ToCategoryResponses(ctx, categories),
		Total:      productTotal,
	}, nil
}

// Suggestions returns autocomplete matches across localized catalog text.
func (q *Queries) Suggestions(ctx context.Context, query string, limit int) (*dto.SuggestionsResponse, error) {
	if query == "" {
		return &dto.SuggestionsResponse{}, nil
	}
	limitPerType := limit/3 + 1

	normalized := i18n.NormalizeSearchQuery(query)
	like := "%" + normalized + "%"

	var suggestions []dto.SuggestionItem

	products, err := q.repo.FindProductsByCatalogQuery(query, limitPerType)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	for _, p := range products {
		nameMap := dto.DecodeCatalogNameI18n(p.NameI18n, p.Name)
		var img *string
		if len(p.Images) > 0 {
			img = &p.Images[0]
		}
		suggestions = append(suggestions, dto.SuggestionItem{
			Type:  "product",
			ID:    &p.ID,
			Name:  nameMap.Resolve(ctx, p.Name),
			Slug:  p.Slug,
			Image: img,
			Price: &p.Price,
		})
	}

	stores, err := q.repo.FindStoresByNameLike(like, limitPerType)
	if err != nil {
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

	categories, err := q.repo.FindCategoriesByCatalogQuery(query, limitPerType)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	for _, cat := range categories {
		nameMap := dto.DecodeCatalogNameI18n(cat.NameI18n, cat.Name)
		suggestions = append(suggestions, dto.SuggestionItem{
			Type: "category",
			ID:   &cat.ID,
			Name: nameMap.Resolve(ctx, cat.Name),
			Slug: cat.Slug,
		})
	}

	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}
	return &dto.SuggestionsResponse{Suggestions: suggestions}, nil
}

// Trending returns the most searched queries from the last 7 days.
func (q *Queries) Trending(limit int) ([]dto.TrendingSearch, error) {
	results, err := q.repo.TrendingQueries(limit)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return results, nil
}

// Commands orchestrates search write use cases.
type Commands struct {
	repo *postgres.SearchRepository
}

// NewCommands creates search command use cases.
func NewCommands(repo *postgres.SearchRepository) *Commands {
	return &Commands{repo: repo}
}

// LogSearch records a search query for trending analytics.
func (c *Commands) LogSearch(query string, userID *uint) error {
	if query == "" {
		return nil
	}
	log := models.SearchLog{
		Query:  query,
		UserID: userID,
	}
	return c.repo.CreateSearchLog(&log)
}
