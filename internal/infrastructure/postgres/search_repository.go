package postgres

import (
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

const searchFacetLimit = 10

// SearchRepository runs catalog search queries with GORM.
type SearchRepository struct {
	db *gorm.DB
}

// NewSearchRepository creates a GORM-backed search repository.
func NewSearchRepository(db *gorm.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// productSearchWhere matches storefront products by FTS, indexed document, and raw catalog fields.
// Falls back to name/description/sku/tags so seeded rows without search_document still match.
func productSearchWhere(query string) (string, []any) {
	normalized := i18n.NormalizeSearchQuery(query)
	if normalized == "" {
		return "1=0", nil
	}
	like := "%" + normalized + "%"
	return `(
		products.search_vector @@ plainto_tsquery('simple', ?)
		OR products.search_vector @@ websearch_to_tsquery('simple', ?)
		OR COALESCE(products.search_document, '') ILIKE ?
		OR products.name ILIKE ?
		OR COALESCE(products.description, '') ILIKE ?
		OR COALESCE(products.sku, '') ILIKE ?
		OR COALESCE(array_to_string(products.tags, ' '), '') ILIKE ?
		OR similarity(COALESCE(NULLIF(products.search_document, ''), products.name), ?) > 0.2
		OR word_similarity(?, COALESCE(NULLIF(products.search_document, ''), products.name)) > 0.35
	)`, []any{normalized, normalized, like, like, like, like, like, normalized, normalized}
}

// catalogSearchWhere builds a qualified FTS clause for categories (and similar catalog tables).
func catalogSearchWhere(table string, query string) (string, []any) {
	normalized := i18n.NormalizeSearchQuery(query)
	if normalized == "" {
		return "1=0", nil
	}
	like := "%" + normalized + "%"
	return fmt.Sprintf(`(
		%s.search_vector @@ plainto_tsquery('simple', ?)
		OR %s.search_vector @@ websearch_to_tsquery('simple', ?)
		OR COALESCE(%s.search_document, '') ILIKE ?
		OR %s.name ILIKE ?
		OR COALESCE(%s.description, '') ILIKE ?
		OR similarity(COALESCE(NULLIF(%s.search_document, ''), %s.name), ?) > 0.2
		OR word_similarity(?, COALESCE(NULLIF(%s.search_document, ''), %s.name)) > 0.35
	)`, table, table, table, table, table, table, table, table, table),
		[]any{normalized, normalized, like, like, like, normalized, normalized}
}

func storeSearchWhere(query string) (string, []any) {
	normalized := i18n.NormalizeSearchQuery(query)
	if normalized == "" {
		return "1=0", nil
	}
	like := "%" + normalized + "%"
	return `(
		search_vector @@ plainto_tsquery('english', ?)
		OR search_vector @@ websearch_to_tsquery('english', ?)
		OR name ILIKE ?
		OR COALESCE(description, '') ILIKE ?
	)`, []any{normalized, normalized, like, like}
}

func (r *SearchRepository) applyProductFilters(query *gorm.DB, req dto.SearchRequest) *gorm.DB {
	query = query.Where("status = ?", constants.ProductStatusActive)

	if req.Query != "" {
		clause, args := productSearchWhere(req.Query)
		query = query.Where(clause, args...)
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

func (r *SearchRepository) applyProductSort(query *gorm.DB, req dto.SearchRequest) *gorm.DB {
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
			normalized := i18n.NormalizeSearchQuery(req.Query)
			return query.Order(gorm.Expr(
				`ts_rank(products.search_vector, plainto_tsquery('simple', ?)) DESC,
				 word_similarity(?, COALESCE(NULLIF(products.search_document, ''), products.name)) DESC`,
				normalized, normalized,
			))
		}
		return query.Order("created_at DESC")
	}
}

// CountProducts counts products matching search filters.
func (r *SearchRepository) CountProducts(req dto.SearchRequest) (int64, error) {
	var total int64
	err := r.applyProductFilters(r.db.Model(&models.Product{}), req).Count(&total).Error
	return total, err
}

// FindProducts returns paginated products matching search filters.
func (r *SearchRepository) FindProducts(req dto.SearchRequest) ([]*models.Product, error) {
	findQuery := r.applyProductFilters(r.db.Model(&models.Product{}), req)
	var products []*models.Product
	err := r.applyProductSort(findQuery, req).
		Limit(req.Limit).
		Offset(req.Offset).
		Preload("Category").
		Find(&products).Error
	return products, err
}

// FindStoresByQuery returns stores matching a text query.
func (r *SearchRepository) FindStoresByQuery(query string) ([]*models.Store, error) {
	clause, args := storeSearchWhere(query)
	var stores []*models.Store
	err := r.db.Where(clause, args...).
		Limit(searchFacetLimit).
		Preload("Categories").
		Find(&stores).Error
	return stores, err
}

// FindCategoriesByQuery returns categories matching a text query.
func (r *SearchRepository) FindCategoriesByQuery(query string) ([]*models.Category, error) {
	clause, args := catalogSearchWhere("categories", query)
	normalized := i18n.NormalizeSearchQuery(query)

	var categories []*models.Category
	err := r.db.Where(clause, args...).
		Order(gorm.Expr(
			`ts_rank(categories.search_vector, plainto_tsquery('simple', ?)) DESC,
			 word_similarity(?, COALESCE(NULLIF(categories.search_document, ''), categories.name)) DESC`,
			normalized, normalized,
		)).
		Limit(searchFacetLimit).
		Find(&categories).Error
	return categories, err
}

// FindProductsByCatalogQuery returns products for autocomplete suggestions.
func (r *SearchRepository) FindProductsByCatalogQuery(query string, limit int) ([]models.Product, error) {
	clause, args := productSearchWhere(query)
	var products []models.Product
	err := r.db.Where("status = ?", constants.ProductStatusActive).
		Where(clause, args...).
		Limit(limit).
		Find(&products).Error
	return products, err
}

// FindStoresByNameLike returns stores matching a name pattern.
func (r *SearchRepository) FindStoresByNameLike(like string, limit int) ([]models.Store, error) {
	var stores []models.Store
	err := r.db.Where("name ILIKE ?", like).Limit(limit).Find(&stores).Error
	return stores, err
}

// FindCategoriesByCatalogQuery returns categories for autocomplete suggestions.
func (r *SearchRepository) FindCategoriesByCatalogQuery(query string, limit int) ([]models.Category, error) {
	clause, args := catalogSearchWhere("categories", query)
	var categories []models.Category
	err := r.db.Where(clause, args...).Limit(limit).Find(&categories).Error
	return categories, err
}

// TrendingQueries returns popular search queries from the last 7 days.
func (r *SearchRepository) TrendingQueries(limit int) ([]dto.TrendingSearch, error) {
	var results []dto.TrendingSearch
	err := r.db.Model(&models.SearchLog{}).
		Select("query, COUNT(*) as count").
		Where("created_at > NOW() - INTERVAL '7 days'").
		Group("query").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error
	return results, err
}

// CreateSearchLog inserts a search log row.
func (r *SearchRepository) CreateSearchLog(log *models.SearchLog) error {
	return r.db.Create(log).Error
}
