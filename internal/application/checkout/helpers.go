package checkout

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	appcart "github.com/alireza-akbarzadeh/luxe/internal/application/cart"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	domaincart "github.com/alireza-akbarzadeh/luxe/internal/domain/cart"
	domaincheckout "github.com/alireza-akbarzadeh/luxe/internal/domain/checkout"
	"gorm.io/gorm"
)

// CartSubtotal sums line totals without mutating stock.
func CartSubtotal(cartItems []models.CartItem) float64 {
	var subtotal float64
	for _, item := range cartItems {
		subtotal += item.Price * float64(item.Quantity)
	}
	return subtotal
}

// ValidateCartForCheckout runs domain rules on a loaded cart model.
func ValidateCartForCheckout(cartDomain *domaincart.Service, userID uint, cart *models.Cart) error {
	domainItems := make([]domaincart.Item, 0, len(cart.Items))
	for _, item := range cart.Items {
		domainItems = append(domainItems, domaincart.Item{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitCents: int64(item.Price * 100),
		})
	}
	return cartDomain.ValidateCheckout(domaincart.Cart{UserID: &userID, Items: domainItems})
}

// ValidateCheckoutInput runs checkout domain validation.
func ValidateCheckoutInput(checkoutDomain *domaincheckout.Service, in domaincheckout.CheckoutInput) error {
	return checkoutDomain.Validate(in)
}

// LoadActiveCart loads the user's active cart for checkout or returns a domain error.
func LoadActiveCart(ctx context.Context, checkoutRepo *postgres.CheckoutRepository, userID uint) (*models.Cart, error) {
	cart, err := checkoutRepo.FindActiveCartForCheckout(ctx, userID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, utils.ErrBadRequest("cart is empty")
		}
		return nil, utils.ErrInternal(err)
	}
	if len(cart.Items) == 0 {
		return nil, utils.ErrBadRequest("cart is empty")
	}
	return cart, nil
}

// ReserveCartStock locks product rows, validates stock, and decrements inventory.
type StockDelta struct {
	Product        models.Product
	QuantityBefore int
	QuantityAfter  int
}

// ProductStockAvailable delegates to cart application stock rules.
func ProductStockAvailable(product models.Product, quantity int) bool {
	return appcart.ProductStockAvailable(product, quantity)
}

// ShouldDecrementProductStock reports whether inventory should be decremented.
func ShouldDecrementProductStock(product models.Product) bool {
	return product.TrackInventory
}

// ReserveCartStock validates and reserves stock inside a transaction.
func ReserveCartStock(
	ctx context.Context,
	tx *gorm.DB,
	orderID uint,
	cartItems []models.CartItem,
	decrement func(ctx context.Context, tx *gorm.DB, orderID, productID uint, qty int) (StockDelta, error),
) ([]StockDelta, error) {
	var changes []StockDelta
	for _, item := range cartItems {
		var product models.Product
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&product, item.ProductID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, utils.ErrBadRequest("product not found")
			}
			return nil, utils.ErrInternal(err)
		}
		if !ProductStockAvailable(product, item.Quantity) {
			return nil, utils.ErrBadRequest(
				fmt.Sprintf("insufficient stock for product: %s", product.Name))
		}
		if ShouldDecrementProductStock(product) {
			if decrement != nil {
				change, err := decrement(ctx, tx, orderID, item.ProductID, item.Quantity)
				if err != nil {
					return nil, err
				}
				changes = append(changes, change)
				continue
			}
			product.Stock -= item.Quantity
			if err := tx.Save(&product).Error; err != nil {
				return nil, utils.ErrInternal(err)
			}
		}
	}
	return changes, nil
}

// BuildOrderModel prepares a new order row for insert.
func BuildOrderModel(userID uint, totalAmount float64, addressID uint, orderNumber string) *models.Order {
	return &models.Order{
		UserID:            userID,
		OrderNumber:       orderNumber,
		Status:            constants.OrderStatusPending,
		TotalAmount:       totalAmount,
		Currency:          "USD",
		ShippingAddressID: &addressID,
		BillingAddressID:  &addressID,
	}
}

// ShipmentUpdate holds fields applied when marking a shipment shipped.
type ShipmentUpdate struct {
	Status         string
	TrackingNumber string
	ShippedAt      time.Time
}

// GenerateOrderNumber builds a unique order number for a user checkout.
func GenerateOrderNumber(userID uint) string {
	return fmt.Sprintf("ORD-%d-%d", userID, time.Now().UnixNano())
}

// StripeEnabled reports whether Stripe checkout is configured.
func StripeEnabled(cfg *config.Config) bool {
	return cfg != nil && cfg.Stripe.Enabled
}
