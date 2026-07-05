package pdp

import (
	"context"
	"errors"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// Queries orchestrates PDP read use cases.
type Queries struct {
	repo *postgres.PdpRepository
}

// NewQueries creates PDP query use cases.
func NewQueries(repo *postgres.PdpRepository) *Queries {
	return &Queries{repo: repo}
}

// GetPriceHistory returns price history points for a product.
func (q *Queries) GetPriceHistory(ctx context.Context, productID uint, days int) ([]dto.PriceHistoryPoint, error) {
	if days <= 0 {
		days = 90
	}
	if days > 365 {
		days = 365
	}
	since := time.Now().AddDate(0, 0, -days)

	rows, err := q.repo.ListPriceHistory(ctx, productID, since)
	if err != nil {
		return nil, err
	}

	points := make([]dto.PriceHistoryPoint, len(rows))
	for i, row := range rows {
		points[i] = dto.PriceHistoryPoint{
			RecordedAt:     row.RecordedAt,
			Price:          row.Price,
			CompareAtPrice: row.CompareAtPrice,
		}
	}
	return points, nil
}

// GetStockHeatmap reconstructs daily availability for PDP transparency charts.
func (q *Queries) GetStockHeatmap(ctx context.Context, product *models.Product, days int) (dto.StockHeatmapData, error) {
	if product == nil {
		return dto.StockHeatmapData{}, gorm.ErrRecordNotFound
	}

	days = normalizeStockHeatmapDays(days)
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(days - 1))

	adjustments, err := q.repo.ListInventoryAdjustmentsSince(ctx, product.ID, start)
	if err != nil {
		return dto.StockHeatmapData{}, err
	}

	prior, priorErr := q.repo.GetLatestInventoryAdjustmentBefore(ctx, product.ID, start)
	if priorErr == nil && prior != nil {
		adjustments = append([]models.InventoryAdjustment{*prior}, adjustments...)
	} else if priorErr != nil && !errors.Is(priorErr, gorm.ErrRecordNotFound) {
		return dto.StockHeatmapData{}, priorErr
	}

	return buildStockHeatmap(*product, adjustments, days), nil
}

// GetAlternatives loads same-barcode products from other stores.
func (q *Queries) GetAlternatives(ctx context.Context, product *models.Product, limit int) ([]*models.Product, error) {
	if product.Barcode == "" {
		return []*models.Product{}, nil
	}
	if limit <= 0 {
		limit = 6
	}
	if limit > 20 {
		limit = 20
	}
	return q.repo.FindAlternativesByBarcode(ctx, product.Barcode, product.ID, product.StoreID, limit)
}

// IsStockSubscribed reports whether the user has an active subscription.
func (q *Queries) IsStockSubscribed(ctx context.Context, userID, productID uint) (bool, error) {
	count, err := q.repo.CountActiveStockSubscription(ctx, userID, productID)
	return count > 0, err
}

// ListQuestionsByUser returns paginated product Q&A asked by a user.
func (q *Queries) ListQuestionsByUser(ctx context.Context, userID uint, limit, offset int) ([]models.ProductQuestion, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	total, err := q.repo.CountQuestionsByUser(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	questions, err := q.repo.ListQuestionsByUser(ctx, userID, limit, offset)
	return questions, total, err
}

// ListQuestions returns paginated questions for a product.
func (q *Queries) ListQuestions(ctx context.Context, productID uint, limit, offset int) ([]models.ProductQuestion, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	total, err := q.repo.CountQuestions(ctx, productID)
	if err != nil {
		return nil, 0, err
	}
	questions, err := q.repo.ListQuestions(ctx, productID, limit, offset)
	return questions, total, err
}

// GetProductWithStore loads a product with store for auto-reply.
func (q *Queries) GetProductWithStore(ctx context.Context, productID uint) (*models.Product, error) {
	return q.repo.GetProductWithStore(ctx, productID)
}

// GetQuestionWithProductStore loads a question with product and store.
func (q *Queries) GetQuestionWithProductStore(ctx context.Context, questionID uint) (*models.ProductQuestion, error) {
	return q.repo.GetQuestionWithProductStore(ctx, questionID)
}

// ToAlternativeResponses maps products to alternative DTOs.
func ToAlternativeResponses(ctx context.Context, alternatives []*models.Product) []dto.ProductAlternativeResponse {
	result := make([]dto.ProductAlternativeResponse, 0, len(alternatives))
	for _, alt := range alternatives {
		item := dto.ProductAlternativeResponse{
			ProductResponse: dto.ToProductResponse(ctx, *alt),
		}
		if alt.Store != nil && alt.Store.ID != 0 {
			item.StoreName = alt.Store.Name
			item.StoreSlug = alt.Store.Slug
			item.StoreLogo = alt.Store.LogoURL
			item.StoreRating = alt.Store.Rating
		}
		result = append(result, item)
	}
	return result
}

// ListActiveSubscriptions returns active stock notifications for a product.
func (q *Queries) ListActiveSubscriptions(ctx context.Context, productID uint) ([]models.StockNotification, error) {
	return q.repo.ListActiveStockSubscriptions(ctx, productID)
}

// ActiveSubscriptionStatus is the active waitlist status constant.
const ActiveSubscriptionStatus = constants.StockNotificationStatusActive
