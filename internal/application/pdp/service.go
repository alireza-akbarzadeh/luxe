package pdp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	appai "github.com/alireza-akbarzadeh/luxe/internal/application/ai"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Notifier sends in-app notifications to users.
type Notifier interface {
	CreateNotification(userID uint, notificationType, title, message string, data interface{}) error
}

// ProductReader loads products for PDP flows.
type ProductReader interface {
	GetByID(id uint) (*models.Product, error)
}

// Service orchestrates product detail page use cases.
type Service struct {
	notifier   Notifier
	products   ProductReader
	aiSvc      *appai.Service
	commands   *Commands
	queries    *Queries
}

// NewService wires PDP commands and queries.
func NewService(
	db *gorm.DB,
	notifier Notifier,
	products ProductReader,
	aiSvc *appai.Service,
) *Service {
	repo := postgres.NewPdpRepository(db)
	return &Service{
		notifier: notifier,
		products: products,
		aiSvc:    aiSvc,
		commands: NewCommands(repo),
		queries:  NewQueries(repo),
	}
}

func (s *Service) RecordPriceSnapshot(product *models.Product) error {
	return s.commands.RecordPriceSnapshot(context.Background(), product)
}

func (s *Service) GetPriceHistory(productID uint, days int) ([]dto.PriceHistoryPoint, error) {
	points, err := s.queries.GetPriceHistory(context.Background(), productID, days)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return points, nil
}

// GetPriceHistoryCtx returns price snapshots for AI and API callers with request context.
func (s *Service) GetPriceHistoryCtx(ctx context.Context, productID uint, days int) ([]dto.PriceHistoryPoint, error) {
	return s.queries.GetPriceHistory(ctx, productID, days)
}

func (s *Service) GetStockHeatmap(ctx context.Context, productID uint, days int) (dto.StockHeatmapData, error) {
	product, err := s.products.GetByID(productID)
	if err != nil {
		return dto.StockHeatmapData{}, err
	}
	data, err := s.queries.GetStockHeatmap(ctx, product, days)
	if err != nil {
		return dto.StockHeatmapData{}, utils.ErrInternal(err)
	}
	return data, nil
}

func (s *Service) GetProductTimeline(ctx context.Context, productID uint, days int) (dto.ProductTimelineData, error) {
	product, err := s.products.GetByID(productID)
	if err != nil {
		return dto.ProductTimelineData{}, err
	}
	data, err := s.queries.GetProductTimeline(ctx, product, days)
	if err != nil {
		return dto.ProductTimelineData{}, utils.ErrInternal(err)
	}
	return data, nil
}

func (s *Service) GetAlternatives(ctx context.Context, productID uint, limit int) ([]dto.ProductAlternativeResponse, error) {
	product, err := s.products.GetByID(productID)
	if err != nil {
		return nil, err
	}

	alternatives, err := s.queries.GetAlternatives(ctx, product, limit)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return ToAlternativeResponses(ctx, alternatives), nil
}

func (s *Service) SubscribeStockNotification(userID, productID uint) error {
	product, err := s.products.GetByID(productID)
	if err != nil {
		return err
	}

	err = s.commands.SubscribeStockNotification(context.Background(), userID, productID, product.Stock > 0)
	if err != nil {
		if errors.Is(err, ErrProductInStock) {
			return utils.ErrBadRequest("product is currently in stock")
		}
		if errors.Is(err, ErrAlreadySubscribed) {
			return utils.ErrBadRequest("you are already subscribed")
		}
		return utils.ErrInternal(err)
	}
	return nil
}

func (s *Service) UnsubscribeStockNotification(userID, productID uint) error {
	rows, err := s.commands.UnsubscribeStockNotification(context.Background(), userID, productID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("subscription not found")
	}
	return nil
}

func (s *Service) IsStockSubscribed(userID, productID uint) (bool, error) {
	return s.queries.IsStockSubscribed(context.Background(), userID, productID)
}

func (s *Service) NotifyBackInStock(productID uint, productName, productSlug string) error {
	subs, err := s.queries.ListActiveSubscriptions(context.Background(), productID)
	if err != nil {
		return err
	}
	for _, sub := range subs {
		_ = s.notifier.CreateNotification(
			sub.UserID,
			"back_in_stock",
			"Back in stock",
			fmt.Sprintf("%s is available again.", productName),
			map[string]interface{}{
				"product_id":   productID,
				"product_slug": productSlug,
				"product_name": productName,
			},
		)
	}
	return s.commands.MarkSubscriptionsSent(context.Background(), productID)
}

func (s *Service) ListQuestions(productID uint, limit, offset int) ([]models.ProductQuestion, int64, error) {
	return s.queries.ListQuestions(context.Background(), productID, limit, offset)
}

func (s *Service) ListQuestionsByUser(ctx context.Context, userID uint, limit, offset int) ([]models.ProductQuestion, int64, error) {
	return s.queries.ListQuestionsByUser(ctx, userID, limit, offset)
}

func (s *Service) CreateQuestion(userID, productID uint, body string) (*models.ProductQuestion, error) {
	if _, err := s.products.GetByID(productID); err != nil {
		return nil, err
	}

	question, err := s.commands.CreateQuestion(context.Background(), userID, productID, body)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	s.maybeAutoReply(question)
	return question, nil
}

func (s *Service) maybeAutoReply(question *models.ProductQuestion) {
	ctx := context.Background()
	product, err := s.queries.GetProductWithStore(ctx, question.ProductID)
	if err != nil {
		return
	}

	reply := BuildAIReply(product, question.Body)
	if s.aiSvc != nil && s.aiSvc.Enabled() {
		if aiReply, err := s.aiSvc.ReplyToQuestion(ctx, product, question.Body); err == nil && strings.TrimSpace(aiReply) != "" {
			reply = aiReply
		}
	}

	ownerID := uint(1)
	if product.Store != nil && product.Store.UserID != nil && *product.Store.UserID > 0 {
		ownerID = *product.Store.UserID
	}

	_ = s.commands.CreateAutoReply(ctx, question.ID, ownerID, reply)
}

func (s *Service) CreateAnswer(userID, questionID uint, body string) (*models.ProductAnswer, error) {
	question, err := s.queries.GetQuestionWithProductStore(context.Background(), questionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("question not found")
		}
		return nil, utils.ErrInternal(err)
	}

	isStoreReply := false
	if question.Product.Store != nil && question.Product.Store.UserID != nil {
		isStoreReply = *question.Product.Store.UserID == userID
	}

	answer, err := s.commands.CreateAnswer(context.Background(), userID, questionID, body, isStoreReply)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return answer, nil
}

func (s *Service) ListDiscussions(productID uint, limit, offset int) ([]models.ProductDiscussion, int64, error) {
	return s.queries.ListDiscussions(context.Background(), productID, limit, offset)
}

func (s *Service) CreateDiscussion(userID, productID uint, title, body string) (*models.ProductDiscussion, error) {
	if _, err := s.products.GetByID(productID); err != nil {
		return nil, err
	}
	discussion, err := s.commands.CreateDiscussion(context.Background(), userID, productID, title, body)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return discussion, nil
}

func (s *Service) CreateDiscussionReply(userID, productID, discussionID uint, body string) (*models.ProductDiscussionReply, error) {
	discussion, err := s.queries.GetDiscussionWithProduct(context.Background(), discussionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("discussion not found")
		}
		return nil, utils.ErrInternal(err)
	}
	if discussion.ProductID != productID {
		return nil, utils.ErrNotFound("discussion not found")
	}
	reply, err := s.commands.CreateDiscussionReply(context.Background(), userID, discussionID, body)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return reply, nil
}
