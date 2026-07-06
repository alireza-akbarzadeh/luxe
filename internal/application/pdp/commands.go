package pdp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// Commands orchestrates PDP write use cases.
type Commands struct {
	repo *postgres.PdpRepository
}

// NewCommands creates PDP command use cases.
func NewCommands(repo *postgres.PdpRepository) *Commands {
	return &Commands{repo: repo}
}

// RecordPriceSnapshot inserts a price history row for a product.
func (c *Commands) RecordPriceSnapshot(ctx context.Context, product *models.Product) error {
	if product == nil || product.ID == 0 {
		return nil
	}
	row := models.ProductPriceHistory{
		ProductID:      product.ID,
		Price:          product.Price,
		CompareAtPrice: product.CompareAtPrice,
		RecordedAt:     time.Now(),
	}
	return c.repo.CreatePriceSnapshot(ctx, &row)
}

// SubscribeStockNotification creates or reactivates a back-in-stock subscription.
func (c *Commands) SubscribeStockNotification(ctx context.Context, userID, productID uint, inStock bool) error {
	if inStock {
		return ErrProductInStock
	}

	existing, err := c.repo.FindStockNotification(ctx, userID, productID)
	if err == nil {
		if existing.Status == constants.StockNotificationStatusActive {
			return ErrAlreadySubscribed
		}
		existing.Status = constants.StockNotificationStatusActive
		existing.NotifiedAt = nil
		return c.repo.SaveStockNotification(ctx, existing)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return c.repo.CreateStockNotification(ctx, &models.StockNotification{
		UserID:    userID,
		ProductID: productID,
		Status:    constants.StockNotificationStatusActive,
	})
}

// UnsubscribeStockNotification cancels an active subscription.
func (c *Commands) UnsubscribeStockNotification(ctx context.Context, userID, productID uint) (int64, error) {
	return c.repo.CancelStockNotification(ctx, userID, productID)
}

// MarkSubscriptionsSent marks active subscriptions as sent.
func (c *Commands) MarkSubscriptionsSent(ctx context.Context, productID uint) error {
	subs, err := c.repo.ListActiveStockSubscriptions(ctx, productID)
	if err != nil {
		return err
	}
	now := time.Now()
	for i := range subs {
		subs[i].Status = "sent"
		subs[i].NotifiedAt = &now
		if err := c.repo.SaveStockNotification(ctx, &subs[i]); err != nil {
			return err
		}
	}
	return nil
}

// CreateQuestion inserts a product question.
func (c *Commands) CreateQuestion(ctx context.Context, userID, productID uint, body string) (*models.ProductQuestion, error) {
	question := &models.ProductQuestion{
		ProductID: productID,
		UserID:    userID,
		Body:      strings.TrimSpace(body),
	}
	if err := c.repo.CreateQuestion(ctx, question); err != nil {
		return nil, err
	}
	return c.repo.GetQuestionByID(ctx, question.ID)
}

// CreateAutoReply inserts an AI/store auto-reply for a question.
func (c *Commands) CreateAutoReply(ctx context.Context, questionID, ownerID uint, reply string) error {
	answer := models.ProductAnswer{
		QuestionID:   questionID,
		UserID:       ownerID,
		Body:         reply,
		IsAIReply:    true,
		IsStoreReply: false,
	}
	return c.repo.CreateAnswer(ctx, &answer)
}

// CreateAnswer inserts a manual answer to a question.
func (c *Commands) CreateAnswer(ctx context.Context, userID, questionID uint, body string, isStoreReply bool) (*models.ProductAnswer, error) {
	answer := &models.ProductAnswer{
		QuestionID:   questionID,
		UserID:       userID,
		Body:         strings.TrimSpace(body),
		IsStoreReply: isStoreReply,
	}
	if err := c.repo.CreateAnswer(ctx, answer); err != nil {
		return nil, err
	}
	return c.repo.GetAnswerByID(ctx, answer.ID)
}

// CreateDiscussion inserts a community discussion thread on a product.
func (c *Commands) CreateDiscussion(ctx context.Context, userID, productID uint, title, body string) (*models.ProductDiscussion, error) {
	discussion := &models.ProductDiscussion{
		ProductID: productID,
		UserID:    userID,
		Title:     strings.TrimSpace(title),
		Body:      strings.TrimSpace(body),
	}
	if err := c.repo.CreateDiscussion(ctx, discussion); err != nil {
		return nil, err
	}
	return c.repo.GetDiscussionByID(ctx, discussion.ID)
}

// CreateDiscussionReply inserts a reply on a discussion thread.
func (c *Commands) CreateDiscussionReply(ctx context.Context, userID, discussionID uint, body string) (*models.ProductDiscussionReply, error) {
	reply := &models.ProductDiscussionReply{
		DiscussionID: discussionID,
		UserID:       userID,
		Body:         strings.TrimSpace(body),
	}
	if err := c.repo.CreateDiscussionReply(ctx, reply); err != nil {
		return nil, err
	}
	return c.repo.GetDiscussionReplyByID(ctx, reply.ID)
}

var (
	ErrProductInStock    = errors.New("product is currently in stock")
	ErrAlreadySubscribed = errors.New("already subscribed")
)

// BuildAIReply builds a fallback reply when AI is unavailable.
func BuildAIReply(product *models.Product, question string) string {
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
