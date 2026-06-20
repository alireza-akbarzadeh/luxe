package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

type PdpServiceInterface interface {
	RecordPriceSnapshot(product *models.Product) error
	GetPriceHistory(productID uint, days int) ([]dto.PriceHistoryPoint, error)
	GetAlternatives(productID uint, limit int) ([]dto.ProductAlternativeResponse, error)
	SubscribeStockNotification(userID, productID uint) error
	UnsubscribeStockNotification(userID, productID uint) error
	IsStockSubscribed(userID, productID uint) (bool, error)
	NotifyBackInStock(productID uint, productName, productSlug string) error
	ListQuestions(productID uint, limit, offset int) ([]models.ProductQuestion, int64, error)
	CreateQuestion(userID, productID uint, body string) (*models.ProductQuestion, error)
	CreateAnswer(userID, questionID uint, body string) (*models.ProductAnswer, error)
}

type pdpService struct {
	db              *gorm.DB
	notificationSvc NotificationServiceInterface
	productSvc      ProductServiceInterface
}

func NewPdpService(db *gorm.DB, notificationSvc NotificationServiceInterface, productSvc ProductServiceInterface) PdpServiceInterface {
	return &pdpService{db: db, notificationSvc: notificationSvc, productSvc: productSvc}
}

func (s *pdpService) RecordPriceSnapshot(product *models.Product) error {
	if product == nil || product.ID == 0 {
		return nil
	}
	row := models.ProductPriceHistory{
		ProductID:      product.ID,
		Price:          product.Price,
		CompareAtPrice: product.CompareAtPrice,
		RecordedAt:     time.Now(),
	}
	return s.db.Create(&row).Error
}

func (s *pdpService) GetPriceHistory(productID uint, days int) ([]dto.PriceHistoryPoint, error) {
	if days <= 0 {
		days = 90
	}
	if days > 365 {
		days = 365
	}
	since := time.Now().AddDate(0, 0, -days)

	var rows []models.ProductPriceHistory
	err := s.db.Where("product_id = ? AND recorded_at >= ?", productID, since).
		Order("recorded_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, utils.ErrInternal(err)
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

func (s *pdpService) GetAlternatives(productID uint, limit int) ([]dto.ProductAlternativeResponse, error) {
	if limit <= 0 {
		limit = 6
	}
	if limit > 20 {
		limit = 20
	}

	product, err := s.productSvc.GetByID(productID)
	if err != nil {
		return nil, err
	}
	if product.Barcode == "" {
		return []dto.ProductAlternativeResponse{}, nil
	}

	var alternatives []*models.Product
	q := s.db.Preload("Store").Preload("Category").Preload("Brand").Preload("Attributes").
		Where("barcode = ? AND id != ? AND status = ?", product.Barcode, productID, constants.ProductStatusActive)
	if product.StoreID != 0 {
		q = q.Where("store_id != ?", product.StoreID)
	}
	if err := q.Order("price ASC").Limit(limit).Find(&alternatives).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	result := make([]dto.ProductAlternativeResponse, 0, len(alternatives))
	for _, alt := range alternatives {
		item := dto.ProductAlternativeResponse{
			ProductResponse: dto.ToProductResponse(*alt),
			StoreName:       "",
			StoreSlug:       "",
			StoreRating:     0,
		}
		if alt.Store != nil && alt.Store.ID != 0 {
			item.StoreName = alt.Store.Name
			item.StoreSlug = alt.Store.Slug
			item.StoreLogo = alt.Store.LogoURL
			item.StoreRating = alt.Store.Rating
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *pdpService) SubscribeStockNotification(userID, productID uint) error {
	product, err := s.productSvc.GetByID(productID)
	if err != nil {
		return err
	}
	if product.Stock > 0 {
		return utils.ErrBadRequest("product is currently in stock")
	}

	var existing models.StockNotification
	err = s.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&existing).Error
	if err == nil {
		if existing.Status == constants.StockNotificationStatusActive {
			return utils.ErrBadRequest("you are already subscribed")
		}
		existing.Status = constants.StockNotificationStatusActive
		existing.NotifiedAt = nil
		return s.db.Save(&existing).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.ErrInternal(err)
	}

	return s.db.Create(&models.StockNotification{
		UserID:    userID,
		ProductID: productID,
		Status:    constants.StockNotificationStatusActive,
	}).Error
}

func (s *pdpService) UnsubscribeStockNotification(userID, productID uint) error {
	result := s.db.Model(&models.StockNotification{}).
		Where("user_id = ? AND product_id = ? AND status = ?", userID, productID, constants.StockNotificationStatusActive).
		Update("status", "cancelled")
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound("subscription not found")
	}
	return nil
}

func (s *pdpService) IsStockSubscribed(userID, productID uint) (bool, error) {
	var count int64
	err := s.db.Model(&models.StockNotification{}).
		Where("user_id = ? AND product_id = ? AND status = ?", userID, productID, constants.StockNotificationStatusActive).
		Count(&count).Error
	return count > 0, err
}

func (s *pdpService) NotifyBackInStock(productID uint, productName, productSlug string) error {
	var subs []models.StockNotification
	if err := s.db.Where("product_id = ? AND status = ?", productID, constants.StockNotificationStatusActive).Find(&subs).Error; err != nil {
		return err
	}
	now := time.Now()
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
		sub.Status = "sent"
		sub.NotifiedAt = &now
		_ = s.db.Save(&sub).Error
	}
	return nil
}

func (s *pdpService) ListQuestions(productID uint, limit, offset int) ([]models.ProductQuestion, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	var total int64
	s.db.Model(&models.ProductQuestion{}).Where("product_id = ?", productID).Count(&total)

	var questions []models.ProductQuestion
	err := s.db.Preload("User").Preload("Answers", func(db *gorm.DB) *gorm.DB {
		return db.Preload("User").Order("created_at ASC")
	}).Where("product_id = ?", productID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&questions).Error
	return questions, total, err
}

func (s *pdpService) CreateQuestion(userID, productID uint, body string) (*models.ProductQuestion, error) {
	if _, err := s.productSvc.GetByID(productID); err != nil {
		return nil, err
	}
	question := &models.ProductQuestion{
		ProductID: productID,
		UserID:    userID,
		Body:      strings.TrimSpace(body),
	}
	if err := s.db.Create(question).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	_ = s.db.Preload("User").First(question, question.ID).Error
	s.maybeAutoReply(question)
	return question, nil
}

func (s *pdpService) maybeAutoReply(question *models.ProductQuestion) {
	var product models.Product
	if err := s.db.Preload("Store").First(&product, question.ProductID).Error; err != nil {
		return
	}

	reply := buildAIReply(&product, question.Body)
	ownerID := uint(1)
	if product.Store != nil && product.Store.UserID != nil && *product.Store.UserID > 0 {
		ownerID = *product.Store.UserID
	}

	answer := models.ProductAnswer{
		QuestionID:   question.ID,
		UserID:       ownerID,
		Body:         reply,
		IsAIReply:    true,
		IsStoreReply: false,
	}
	_ = s.db.Create(&answer).Error
}

func buildAIReply(product *models.Product, question string) string {
	q := strings.ToLower(question)
	storeName := "our store"
	if product.Store != nil && product.Store.Name != "" {
		storeName = product.Store.Name
	}

	switch {
	case strings.Contains(q, "ship") || strings.Contains(q, "delivery"):
		if product.Store != nil && product.Store.ShippingInfo != "" {
			return fmt.Sprintf("Thanks for asking! %s", product.Store.ShippingInfo)
		}
		return fmt.Sprintf("Thanks for your question. %s ships within 2–5 business days. Express options may be available at checkout.", storeName)
	case strings.Contains(q, "return") || strings.Contains(q, "refund"):
		if product.Store != nil && product.Store.ReturnPolicy != "" {
			return product.Store.ReturnPolicy
		}
		return "This item is eligible for returns within 30 days in original condition. Contact us if you need help starting a return."
	case strings.Contains(q, "size") || strings.Contains(q, "fit"):
		return "Sizing can vary by brand. Check the specifications tab for measurements, and feel free to compare with a similar item you already own."
	case strings.Contains(q, "authentic") || strings.Contains(q, "real"):
		return fmt.Sprintf("%s guarantees authenticity on every listing. Each order is quality-checked before dispatch.", storeName)
	default:
		snippet := product.Description
		if len(snippet) > 180 {
			snippet = snippet[:180] + "…"
		}
		if snippet != "" {
			return fmt.Sprintf("Great question. Based on the listing details: %s If you need anything else, the %s team is happy to help.", snippet, storeName)
		}
		return fmt.Sprintf("Thanks for your interest! The %s team will follow up shortly with more details.", storeName)
	}
}

func (s *pdpService) CreateAnswer(userID, questionID uint, body string) (*models.ProductAnswer, error) {
	var question models.ProductQuestion
	if err := s.db.Preload("Product.Store").First(&question, questionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("question not found")
		}
		return nil, utils.ErrInternal(err)
	}

	isStoreReply := false
	if question.Product.Store != nil && question.Product.Store.UserID != nil {
		isStoreReply = *question.Product.Store.UserID == userID
	}

	answer := &models.ProductAnswer{
		QuestionID:   questionID,
		UserID:       userID,
		Body:         strings.TrimSpace(body),
		IsStoreReply: isStoreReply,
	}
	if err := s.db.Create(answer).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	_ = s.db.Preload("User").First(answer, answer.ID).Error
	return answer, nil
}
