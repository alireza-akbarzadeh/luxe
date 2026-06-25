package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	apppdp "github.com/alireza-akbarzadeh/luxe/internal/application/pdp"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

type PdpServiceInterface interface {
	RecordPriceSnapshot(product *models.Product) error
	GetPriceHistory(productID uint, days int) ([]dto.PriceHistoryPoint, error)
	GetAlternatives(ctx context.Context, productID uint, limit int) ([]dto.ProductAlternativeResponse, error)
	SubscribeStockNotification(userID, productID uint) error
	UnsubscribeStockNotification(userID, productID uint) error
	IsStockSubscribed(userID, productID uint) (bool, error)
	NotifyBackInStock(productID uint, productName, productSlug string) error
	ListQuestions(productID uint, limit, offset int) ([]models.ProductQuestion, int64, error)
	CreateQuestion(userID, productID uint, body string) (*models.ProductQuestion, error)
	CreateAnswer(userID, questionID uint, body string) (*models.ProductAnswer, error)
}

type pdpService struct {
	notificationSvc NotificationServiceInterface
	productSvc      ProductServiceInterface
	aiSvc           AiServiceInterface
	commands        *apppdp.Commands
	queries         *apppdp.Queries
}

func NewPdpService(
	db *gorm.DB,
	notificationSvc NotificationServiceInterface,
	productSvc ProductServiceInterface,
	aiSvc AiServiceInterface,
) PdpServiceInterface {
	repo := postgres.NewPdpRepository(db)
	return &pdpService{
		notificationSvc: notificationSvc,
		productSvc:      productSvc,
		aiSvc:           aiSvc,
		commands:        apppdp.NewCommands(repo),
		queries:         apppdp.NewQueries(repo),
	}
}

func (s *pdpService) RecordPriceSnapshot(product *models.Product) error {
	return s.commands.RecordPriceSnapshot(context.Background(), product)
}

func (s *pdpService) GetPriceHistory(productID uint, days int) ([]dto.PriceHistoryPoint, error) {
	points, err := s.queries.GetPriceHistory(context.Background(), productID, days)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return points, nil
}

func (s *pdpService) GetAlternatives(ctx context.Context, productID uint, limit int) ([]dto.ProductAlternativeResponse, error) {
	product, err := s.productSvc.GetByID(productID)
	if err != nil {
		return nil, err
	}

	alternatives, err := s.queries.GetAlternatives(ctx, product, limit)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return apppdp.ToAlternativeResponses(ctx, alternatives), nil
}

func (s *pdpService) SubscribeStockNotification(userID, productID uint) error {
	product, err := s.productSvc.GetByID(productID)
	if err != nil {
		return err
	}

	err = s.commands.SubscribeStockNotification(context.Background(), userID, productID, product.Stock > 0)
	if err != nil {
		if errors.Is(err, apppdp.ErrProductInStock) {
			return utils.ErrBadRequest("product is currently in stock")
		}
		if errors.Is(err, apppdp.ErrAlreadySubscribed) {
			return utils.ErrBadRequest("you are already subscribed")
		}
		return utils.ErrInternal(err)
	}
	return nil
}

func (s *pdpService) UnsubscribeStockNotification(userID, productID uint) error {
	rows, err := s.commands.UnsubscribeStockNotification(context.Background(), userID, productID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("subscription not found")
	}
	return nil
}

func (s *pdpService) IsStockSubscribed(userID, productID uint) (bool, error) {
	return s.queries.IsStockSubscribed(context.Background(), userID, productID)
}

func (s *pdpService) NotifyBackInStock(productID uint, productName, productSlug string) error {
	subs, err := s.queries.ListActiveSubscriptions(context.Background(), productID)
	if err != nil {
		return err
	}
	for _, sub := range subs {
		_ = s.notificationSvc.CreateNotification(
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

func (s *pdpService) ListQuestions(productID uint, limit, offset int) ([]models.ProductQuestion, int64, error) {
	return s.queries.ListQuestions(context.Background(), productID, limit, offset)
}

func (s *pdpService) CreateQuestion(userID, productID uint, body string) (*models.ProductQuestion, error) {
	if _, err := s.productSvc.GetByID(productID); err != nil {
		return nil, err
	}

	question, err := s.commands.CreateQuestion(context.Background(), userID, productID, body)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	s.maybeAutoReply(question)
	return question, nil
}

func (s *pdpService) maybeAutoReply(question *models.ProductQuestion) {
	ctx := context.Background()
	product, err := s.queries.GetProductWithStore(ctx, question.ProductID)
	if err != nil {
		return
	}

	reply := apppdp.BuildAIReply(product, question.Body)
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

func (s *pdpService) CreateAnswer(userID, questionID uint, body string) (*models.ProductAnswer, error) {
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
