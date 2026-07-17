package collection

import (
	"context"
	"sort"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

const (
	defaultCollectionProductsLimit = 12
	maxCollectionProductsLimit     = 48
	hybridDynamicFetchCap          = 240
)

func normalizeLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = defaultCollectionProductsLimit
	}
	if limit > maxCollectionProductsLimit {
		limit = maxCollectionProductsLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func mergeRequestOverrides(filters *dto.ProductListFilters, collection *models.Collection, req *dto.CollectionProductsRequest) {
	if filters.Sort == "" {
		filters.Sort = collection.SortKey
	}
	if filters.Sort == "" {
		filters.Sort = collection.PreviewSort
	}
	if req.Sort != "" {
		filters.Sort = req.Sort
	}
	if collection.PreviewCategoryID != nil && filters.CategoryID == 0 {
		filters.CategoryID = *collection.PreviewCategoryID
	}
	if collection.PreviewIsNew != nil && filters.IsNew == nil {
		filters.IsNew = collection.PreviewIsNew
	}
	if req.CategoryID != 0 {
		filters.CategoryID = req.CategoryID
	}
	if req.MinPrice > 0 {
		filters.MinPrice = req.MinPrice
	}
	if req.MaxPrice > 0 {
		filters.MaxPrice = req.MaxPrice
	}
	if req.MinRating > 0 {
		filters.MinRating = req.MinRating
	}
	if req.Search != "" {
		filters.Search = req.Search
	}
	if req.IsNew != nil {
		filters.IsNew = req.IsNew
	}
	if req.InStock != nil {
		filters.InStock = req.InStock
	}
	if req.OnSale != nil {
		filters.OnSale = req.OnSale
	}
}

func applyRuleCondition(filters *dto.ProductListFilters, condition dto.CollectionRuleCondition) error {
	field := strings.TrimSpace(condition.Field)
	operator := strings.ToLower(strings.TrimSpace(condition.Operator))
	if operator == "" {
		operator = "eq"
	}
	switch field {
	case "category_id":
		if operator == "in" {
			values, ok := toUintSlice(condition.Value)
			if !ok || len(values) == 0 {
				return utils.ErrBadRequest("category_id in requires ids")
			}
			filters.IDs = mergeUintIDs(filters.IDs, values)
			return nil
		}
		if value, ok := toUint(condition.Value); ok && value > 0 {
			if operator == "neq" {
				return nil // product list has no neq category; skip silently for resolution
			}
			filters.CategoryID = value
		}
	case "brand_id":
		if operator == "in" {
			values, ok := toUintSlice(condition.Value)
			if !ok || len(values) == 0 {
				return utils.ErrBadRequest("brand_id in requires ids")
			}
			// Prefer first brand for single-filter path; OR expansion handles multi-brand.
			filters.BrandID = &values[0]
			return nil
		}
		if value, ok := toUint(condition.Value); ok && value > 0 {
			filters.BrandID = &value
		}
	case "min_price":
		if value, ok := toFloat(condition.Value); ok {
			if operator == "lte" {
				if filters.MaxPrice == 0 || value < filters.MaxPrice {
					filters.MaxPrice = value
				}
			} else {
				filters.MinPrice = value
			}
		}
	case "max_price":
		if value, ok := toFloat(condition.Value); ok {
			if operator == "gte" {
				filters.MinPrice = value
			} else {
				filters.MaxPrice = value
			}
		}
	case "min_rating":
		if value, ok := toFloat(condition.Value); ok {
			filters.MinRating = value
		}
	case "is_new":
		if value, ok := toBool(condition.Value); ok {
			filters.IsNew = &value
		}
	case "in_stock":
		if value, ok := toBool(condition.Value); ok {
			filters.InStock = &value
		}
	case "on_sale":
		if value, ok := toBool(condition.Value); ok {
			filters.OnSale = &value
		}
	case "search":
		if value, ok := condition.Value.(string); ok {
			filters.Search = value
		}
	case "sort":
		if value, ok := condition.Value.(string); ok {
			filters.Sort = value
		}
	case "ids":
		if values, ok := toUintSlice(condition.Value); ok {
			filters.IDs = mergeUintIDs(filters.IDs, values)
		} else if value, ok := toUint(condition.Value); ok {
			filters.IDs = mergeUintIDs(filters.IDs, []uint{value})
		}
	}
	return nil
}

func mergeUintIDs(existing, extra []uint) []uint {
	seen := make(map[uint]struct{}, len(existing)+len(extra))
	out := make([]uint, 0, len(existing)+len(extra))
	for _, id := range existing {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	for _, id := range extra {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (s *Service) resolveCollectionProducts(
	ctx context.Context,
	collection *models.Collection,
	req *dto.CollectionProductsRequest,
) (*dto.ProductListData, error) {
	limit, offset := normalizeLimitOffset(req.Limit, req.Offset)
	mode := normalizeMode(collection.Mode, collection.CollectionType)
	rules := unmarshalRules(collection.RulesJSON)
	branches, err := expandRuleBranches(rules)
	if err != nil {
		return nil, err
	}
	for i := range branches {
		mergeRequestOverrides(&branches[i], collection, req)
	}

	switch mode {
	case "manual":
		return s.resolveManualProducts(ctx, collection, branches[0], limit, offset)
	case "hybrid":
		return s.resolveHybridProducts(ctx, collection, branches, limit, offset)
	default:
		return s.resolveDynamicProducts(ctx, branches, limit, offset)
	}
}

func (s *Service) resolveDynamicProducts(
	ctx context.Context,
	branches []dto.ProductListFilters,
	limit int,
	offset int,
) (*dto.ProductListData, error) {
	if len(branches) <= 1 {
		filters := dto.ProductListFilters{Status: "active"}
		if len(branches) == 1 {
			filters = branches[0]
		}
		products, total, err := s.products.ListDetailed(ctx, limit, offset, filters)
		if err != nil {
			return nil, err
		}
		return buildProductListData(ctx, products, total, limit, offset), nil
	}
	return s.resolveUnionBranches(ctx, branches, limit, offset)
}

func (s *Service) resolveUnionBranches(
	ctx context.Context,
	branches []dto.ProductListFilters,
	limit int,
	offset int,
) (*dto.ProductListData, error) {
	merged := make([]*models.Product, 0)
	seen := make(map[uint]struct{})
	sortKey := ""
	for _, filters := range branches {
		if sortKey == "" {
			sortKey = filters.Sort
		}
		products, _, err := s.products.ListDetailed(ctx, hybridDynamicFetchCap, 0, filters)
		if err != nil {
			return nil, err
		}
		for _, product := range products {
			if product == nil {
				continue
			}
			if _, ok := seen[product.ID]; ok {
				continue
			}
			seen[product.ID] = struct{}{}
			merged = append(merged, product)
		}
	}
	sortLoadedProducts(merged, sortKey)
	total := int64(len(merged))
	paged := paginateProducts(merged, limit, offset)
	return buildProductListData(ctx, paged, total, limit, offset), nil
}

func (s *Service) resolveManualProducts(
	ctx context.Context,
	collection *models.Collection,
	filters dto.ProductListFilters,
	limit int,
	offset int,
) (*dto.ProductListData, error) {
	products, err := s.loadManualProducts(ctx, collection)
	if err != nil {
		return nil, err
	}
	filtered := filterLoadedProducts(products, filters)
	total := int64(len(filtered))
	paged := paginateProducts(filtered, limit, offset)
	return buildProductListData(ctx, paged, total, limit, offset), nil
}

func (s *Service) resolveHybridProducts(
	ctx context.Context,
	collection *models.Collection,
	branches []dto.ProductListFilters,
	limit int,
	offset int,
) (*dto.ProductListData, error) {
	dynamicProducts := make([]*models.Product, 0)
	seenDynamic := make(map[uint]struct{})
	for _, filters := range branches {
		batch, _, err := s.products.ListDetailed(ctx, hybridDynamicFetchCap, 0, filters)
		if err != nil {
			return nil, err
		}
		for _, product := range batch {
			if product == nil {
				continue
			}
			if _, ok := seenDynamic[product.ID]; ok {
				continue
			}
			seenDynamic[product.ID] = struct{}{}
			dynamicProducts = append(dynamicProducts, product)
		}
	}
	manualProducts, err := s.loadManualProducts(ctx, collection)
	if err != nil {
		return nil, err
	}

	overrideByProductID := make(map[uint]models.CollectionProduct, len(collection.Products))
	for _, override := range collection.Products {
		overrideByProductID[override.ProductID] = override
	}

	merged := make([]*models.Product, 0, len(dynamicProducts)+len(manualProducts))
	seen := make(map[uint]struct{})
	for _, product := range dynamicProducts {
		override, hasOverride := overrideByProductID[product.ID]
		if hasOverride && override.IsHidden {
			continue
		}
		merged = append(merged, product)
		seen[product.ID] = struct{}{}
	}
	for _, product := range manualProducts {
		override := overrideByProductID[product.ID]
		if override.IsHidden {
			continue
		}
		if _, exists := seen[product.ID]; exists {
			continue
		}
		merged = append(merged, product)
		seen[product.ID] = struct{}{}
	}

	sort.SliceStable(merged, func(i, j int) bool {
		left := overrideByProductID[merged[i].ID]
		right := overrideByProductID[merged[j].ID]
		if left.IsPinned != right.IsPinned {
			return left.IsPinned
		}
		if left.Position != right.Position {
			return left.Position < right.Position
		}
		if left.BoostScore != right.BoostScore {
			return left.BoostScore > right.BoostScore
		}
		return merged[i].ID < merged[j].ID
	})

	total := int64(len(merged))
	paged := paginateProducts(merged, limit, offset)
	return buildProductListData(ctx, paged, total, limit, offset), nil
}

func (s *Service) loadManualProducts(ctx context.Context, collection *models.Collection) ([]*models.Product, error) {
	if len(collection.Products) == 0 {
		return nil, nil
	}
	needsFetch := false
	for _, item := range collection.Products {
		if item.Product == nil {
			needsFetch = true
			break
		}
	}
	if !needsFetch {
		products := make([]*models.Product, 0, len(collection.Products))
		for _, item := range collection.Products {
			products = append(products, item.Product)
		}
		return products, nil
	}

	ids := ProductIDsFromModel(collection)
	loaded, _, err := s.products.ListDetailed(ctx, len(ids), 0, dto.ProductListFilters{
		Status: "active",
		IDs:    ids,
	})
	if err != nil {
		return nil, err
	}
	byID := make(map[uint]*models.Product, len(loaded))
	for _, product := range loaded {
		byID[product.ID] = product
	}
	ordered := make([]*models.Product, 0, len(ids))
	for _, id := range ids {
		if product, ok := byID[id]; ok {
			ordered = append(ordered, product)
		}
	}
	return ordered, nil
}

func filterLoadedProducts(products []*models.Product, filters dto.ProductListFilters) []*models.Product {
	if len(products) == 0 {
		return nil
	}
	filtered := make([]*models.Product, 0, len(products))
	for _, product := range products {
		if product == nil {
			continue
		}
		if filters.CategoryID != 0 && (product.CategoryID == nil || *product.CategoryID != filters.CategoryID) {
			continue
		}
		if filters.BrandID != nil && (product.BrandID == nil || *product.BrandID != *filters.BrandID) {
			continue
		}
		if filters.MinPrice > 0 && product.Price < filters.MinPrice {
			continue
		}
		if filters.MaxPrice > 0 && product.Price > filters.MaxPrice {
			continue
		}
		if filters.MinRating > 0 && product.Rating < filters.MinRating {
			continue
		}
		if filters.IsNew != nil && product.IsNew != *filters.IsNew {
			continue
		}
		if filters.InStock != nil && *filters.InStock && !(product.Stock > 0 || product.AllowBackorder) {
			continue
		}
		if filters.OnSale != nil && *filters.OnSale && (product.CompareAtPrice == nil || *product.CompareAtPrice <= product.Price) {
			continue
		}
		if filters.Search != "" && !matchesSearch(product, filters.Search) {
			continue
		}
		filtered = append(filtered, product)
	}
	sortLoadedProducts(filtered, filters.Sort)
	return filtered
}

func matchesSearch(product *models.Product, term string) bool {
	if term == "" {
		return true
	}
	name := product.Name
	sku := product.SKU
	return containsFold(name, term) || containsFold(sku, term)
}

func containsFold(value, needle string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(needle))
}

func containsInsensitive(value, needle string) bool {
	return containsFold(value, needle)
}

func sortLoadedProducts(products []*models.Product, sortKey string) {
	switch sortKey {
	case "price_asc":
		sort.SliceStable(products, func(i, j int) bool { return products[i].Price < products[j].Price })
	case "price_desc":
		sort.SliceStable(products, func(i, j int) bool { return products[i].Price > products[j].Price })
	case "rating_desc":
		sort.SliceStable(products, func(i, j int) bool { return products[i].Rating > products[j].Rating })
	case "reviews_desc":
		sort.SliceStable(products, func(i, j int) bool { return products[i].ReviewsCount > products[j].ReviewsCount })
	case "newest":
		sort.SliceStable(products, func(i, j int) bool { return products[i].CreatedAt.After(products[j].CreatedAt) })
	}
}

func paginateProducts(products []*models.Product, limit int, offset int) []*models.Product {
	if offset >= len(products) {
		return []*models.Product{}
	}
	end := offset + limit
	if end > len(products) {
		end = len(products)
	}
	return products[offset:end]
}

func buildProductListData(
	ctx context.Context,
	products []*models.Product,
	total int64,
	limit int,
	offset int,
) *dto.ProductListData {
	items := make([]dto.ProductWithLike, 0, len(products))
	for _, product := range products {
		if product == nil {
			continue
		}
		items = append(items, dto.ProductWithLike{
			ProductResponse: dto.ToProductResponse(ctx, *product),
			IsLiked:         false,
		})
	}
	return &dto.ProductListData{
		Products: items,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}
}

func toUint(value any) (uint, bool) {
	switch v := value.(type) {
	case float64:
		return uint(v), true
	case int:
		return uint(v), true
	case uint:
		return v, true
	}
	return 0, false
}

func toFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case uint:
		return float64(v), true
	}
	return 0, false
}

func toBool(value any) (bool, bool) {
	v, ok := value.(bool)
	return v, ok
}

func toUintSlice(value any) ([]uint, bool) {
	items, ok := value.([]any)
	if !ok {
		return nil, false
	}
	result := make([]uint, 0, len(items))
	for _, item := range items {
		parsed, parsedOK := toUint(item)
		if !parsedOK {
			return nil, false
		}
		result = append(result, parsed)
	}
	return result, true
}
