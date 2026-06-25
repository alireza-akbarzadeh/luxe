package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// CheckoutRepository holds checkout-related persistence.
type CheckoutRepository struct {
	db *gorm.DB
}

// NewCheckoutRepository creates a GORM-backed checkout repository.
func NewCheckoutRepository(db *gorm.DB) *CheckoutRepository {
	return &CheckoutRepository{db: db}
}

// FindActiveCartForCheckout loads the user's active cart with items and products.
func (r *CheckoutRepository) FindActiveCartForCheckout(ctx context.Context, userID uint) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, constants.CartStatusActive).
		Preload("Items.Product").
		First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

// ResolveOrCreateAddress finds or creates a shipping address for checkout.
func (r *CheckoutRepository) ResolveOrCreateAddress(ctx context.Context, userID uint, req dto.CheckoutRequest) (*models.Address, error) {
	address := dto.MapAddress(userID, req)
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND address_line1 = ? AND postal_code = ?", userID, req.AddressLine1, req.Zip).
		FirstOrCreate(&address, address).Error
	if err != nil {
		return nil, err
	}
	return &address, nil
}

// CreateOrder inserts an order row inside a transaction.
func (r *CheckoutRepository) CreateOrder(tx *gorm.DB, order *models.Order) error {
	return tx.Create(order).Error
}

// CreateOrderItems copies cart lines into order items inside a transaction.
func (r *CheckoutRepository) CreateOrderItems(tx *gorm.DB, orderID uint, cartItems []models.CartItem) error {
	for _, item := range cartItems {
		oi := &models.OrderItem{
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
		if err := tx.Create(oi).Error; err != nil {
			return err
		}
	}
	return nil
}

// PreloadOrderDetails loads order relations after checkout.
func (r *CheckoutRepository) PreloadOrderDetails(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).
		Preload("Items.Product").
		Preload("User").
		Preload("Payment").
		First(order, order.ID).Error
}

// FindOrderWithPayment loads an order with its payment record.
func (r *CheckoutRepository) FindOrderWithPayment(ctx context.Context, orderID uint) (*models.Order, error) {
	var order models.Order
	if err := r.db.WithContext(ctx).Preload("Payment").First(&order, orderID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateOrderStatus sets order.status.
func (r *CheckoutRepository) UpdateOrderStatus(ctx context.Context, orderID uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

// UpdateOrderStatusTx sets order.status inside a transaction.
func (r *CheckoutRepository) UpdateOrderStatusTx(tx *gorm.DB, orderID uint, status string) error {
	return tx.Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

// FindOrderForCancel loads an order with items and payment for cancellation.
func (r *CheckoutRepository) FindOrderForCancel(ctx context.Context, orderID, userID uint) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Payment").
		Where("id = ? AND user_id = ?", orderID, userID).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindShipmentByOrderID loads the shipment for an order.
func (r *CheckoutRepository) FindShipmentByOrderID(ctx context.Context, orderID uint) (*models.Shipment, error) {
	var shipment models.Shipment
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&shipment).Error; err != nil {
		return nil, err
	}
	return &shipment, nil
}

// SaveShipment persists shipment changes.
func (r *CheckoutRepository) SaveShipment(ctx context.Context, shipment *models.Shipment) error {
	return r.db.WithContext(ctx).Save(shipment).Error
}

// UpdateShipmentFields updates shipment columns by order id.
func (r *CheckoutRepository) UpdateShipmentFields(ctx context.Context, orderID uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Shipment{}).Where("order_id = ?", orderID).Updates(updates).Error
}

// UpdateShipmentStatusTx updates shipment status inside a transaction.
func (r *CheckoutRepository) UpdateShipmentStatusTx(tx *gorm.DB, orderID uint, status string) error {
	return tx.Model(&models.Shipment{}).Where("order_id = ?", orderID).Update("status", status).Error
}

// CancelPendingShipmentsTx cancels non-shipped shipments for an order.
func (r *CheckoutRepository) CancelPendingShipmentsTx(tx *gorm.DB, orderID uint) error {
	return tx.Model(&models.Shipment{}).
		Where("order_id = ? AND status NOT IN ?", orderID,
			[]string{constants.ShipmentStatusShipped, constants.ShipmentStatusDelivered}).
		Update("status", "cancelled").Error
}

// FindPaymentByOrderID loads payment for an order.
func (r *CheckoutRepository) FindPaymentByOrderID(ctx context.Context, orderID uint) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// IsNotFound reports gorm record-not-found errors.
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// Transaction runs fn inside a database transaction.
func (r *CheckoutRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// GetProductForUpdateTx loads a product row with FOR UPDATE inside a transaction.
func (r *CheckoutRepository) GetProductForUpdateTx(tx *gorm.DB, productID uint) (*models.Product, error) {
	var product models.Product
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&product, productID).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// SaveProductTx persists product changes inside a transaction.
func (r *CheckoutRepository) SaveProductTx(tx *gorm.DB, product *models.Product) error {
	return tx.Save(product).Error
}

// RestoreProductStockTx increments product stock inside a transaction.
func (r *CheckoutRepository) RestoreProductStockTx(tx *gorm.DB, productID uint, quantity int) error {
	return tx.Model(&models.Product{}).
		Where("id = ?", productID).
		UpdateColumn("stock", gorm.Expr("stock + ?", quantity)).Error
}

// FindOrderWithPaymentTx loads an order and payment inside a transaction.
func (r *CheckoutRepository) FindOrderWithPaymentTx(tx *gorm.DB, orderID uint) (*models.Order, error) {
	var order models.Order
	if err := tx.Preload("Payment").First(&order, orderID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// MarkOrderPaymentFailedTx sets order and shipment to failed/cancelled states.
func (r *CheckoutRepository) MarkOrderPaymentFailedTx(tx *gorm.DB, orderID uint) error {
	if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Update("status", "payment_failed").Error; err != nil {
		return err
	}
	return tx.Model(&models.Shipment{}).Where("order_id = ?", orderID).Update("status", "cancelled").Error
}

// UpdateWalletPaymentSucceededTx marks a wallet payment as succeeded inside a transaction.
func (r *CheckoutRepository) UpdateWalletPaymentSucceededTx(tx *gorm.DB, paymentID, userID, orderID uint) error {
	return tx.Model(&models.Payment{}).Where("id = ?", paymentID).Updates(map[string]interface{}{
		"status":         constants.PaymentStatusSucceeded,
		"transaction_id": fmt.Sprintf("wallet_%d_%d", userID, orderID),
	}).Error
}

// UpdatePaymentRefundedTx marks a payment as refunded inside a transaction.
func (r *CheckoutRepository) UpdatePaymentRefundedTx(tx *gorm.DB, paymentID uint) error {
	return tx.Model(&models.Payment{}).Where("id = ?", paymentID).
		Update("status", constants.PaymentStatusRefunded).Error
}
