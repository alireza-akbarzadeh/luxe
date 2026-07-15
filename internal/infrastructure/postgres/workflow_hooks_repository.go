package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// WorkflowHooksRepository supports workflow guard/hook side effects (GORM only).
type WorkflowHooksRepository struct {
	db *gorm.DB
}

// NewWorkflowHooksRepository creates a GORM-backed workflow hooks repository.
func NewWorkflowHooksRepository(db *gorm.DB) *WorkflowHooksRepository {
	return &WorkflowHooksRepository{db: db}
}

func (r *WorkflowHooksRepository) GetProductPrice(ctx context.Context, productID uint) (float64, error) {
	var price float64
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Select("price").Where("id = ?", productID).Scan(&price).Error
	return price, err
}

func (r *WorkflowHooksRepository) GetLatestPaymentStatusForOrder(ctx context.Context, orderID uint) (string, error) {
	var status string
	err := r.db.WithContext(ctx).Model(&models.Payment{}).
		Select("status").Where("order_id = ?", orderID).
		Order("id DESC").Limit(1).Scan(&status).Error
	return status, err
}

func (r *WorkflowHooksRepository) GetOrderStatus(ctx context.Context, orderID uint) (string, error) {
	var status string
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Select("status").Where("id = ?", orderID).Scan(&status).Error
	return status, err
}

func (r *WorkflowHooksRepository) SetProductPublishedAt(ctx context.Context, productID uint, publishedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Product{}).
		Where("id = ?", productID).Update("published_at", publishedAt).Error
}

func (r *WorkflowHooksRepository) SetBlogPostPublishedAt(ctx context.Context, postID uint, publishedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.BlogPost{}).
		Where("id = ? AND published_at IS NULL", postID).
		Update("published_at", publishedAt).Error
}

func (r *WorkflowHooksRepository) SetShipmentDeliveredAt(ctx context.Context, shipmentID uint, deliveredAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Shipment{}).
		Where("id = ?", shipmentID).Update("delivered_at", deliveredAt).Error
}

func (r *WorkflowHooksRepository) GetReturnRefundInfo(ctx context.Context, returnID uint) (models.Return, error) {
	var row models.Return
	err := r.db.WithContext(ctx).Model(&models.Return{}).
		Select("order_id", "user_id", "refund_amount").
		Where("id = ?", returnID).First(&row).Error
	return row, err
}

type OrderNotifyRow struct {
	UserID      uint
	OrderNumber string
}

func (r *WorkflowHooksRepository) GetOrderNotifyRow(ctx context.Context, orderID uint) (OrderNotifyRow, error) {
	var row OrderNotifyRow
	err := r.db.WithContext(ctx).Model(&models.Order{}).
		Select("user_id", "order_number").
		Where("id = ?", orderID).Scan(&row).Error
	return row, err
}

func (r *WorkflowHooksRepository) GetUserEmail(ctx context.Context, userID uint) (string, error) {
	var email string
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Select("email").Where("id = ?", userID).Scan(&email).Error
	return email, err
}
