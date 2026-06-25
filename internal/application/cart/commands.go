package cart

import (
	"context"
	"errors"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// AddItemInput is the application-layer add-to-cart command.
type AddItemInput struct {
	ProductID uint
	Quantity  int
}

// UpdateItemInput updates quantity and optional variant fields.
type UpdateItemInput struct {
	Quantity int
	Color    string
	Size     string
}

// Commands orchestrates cart write use cases.
type Commands struct {
	reader Reader
	writer Writer
}

// NewCommands creates cart command use cases.
func NewCommands(reader Reader, writer Writer) *Commands {
	return &Commands{reader: reader, writer: writer}
}

// GetOrCreateCart returns an active cart, creating one when absent.
func (c *Commands) GetOrCreateCart(ctx context.Context, userID uint) (*models.Cart, error) {
	cart, err := c.reader.FindActiveCart(ctx, userID, true)
	if err == nil {
		return cart, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrInternal(err)
	}
	newCart, err := c.writer.CreateActiveCart(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return newCart, nil
}

// AddItem adds or merges a line on the user's active cart.
func (c *Commands) AddItem(ctx context.Context, userID uint, in AddItemInput) (*models.CartItem, error) {
	if in.Quantity <= 0 {
		return nil, utils.ErrBadRequest("quantity must be positive")
	}

	product, err := c.reader.GetProductByID(ctx, in.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}
	if product.Status != constants.ProductStatusActive {
		return nil, utils.ErrBadRequest("product is not available")
	}
	if !ProductStockAvailable(*product, in.Quantity) {
		return nil, utils.ErrBadRequest("insufficient stock")
	}

	cart, err := c.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	cartItem, err := c.reader.FindCartItemByCartAndProduct(ctx, cart.ID, in.ProductID)
	if err == nil {
		newQty := cartItem.Quantity + in.Quantity
		if !ProductStockAvailable(*product, newQty) {
			return nil, utils.ErrBadRequest("insufficient stock for updated quantity")
		}
		cartItem.Quantity = newQty
		if err := c.writer.SaveCartItem(ctx, cartItem); err != nil {
			return nil, utils.ErrInternal(err)
		}
		return cartItem, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrInternal(err)
	}

	newItem := models.CartItem{
		CartID:    cart.ID,
		ProductID: in.ProductID,
		Quantity:  in.Quantity,
		Price:     product.Price,
	}
	if err := c.writer.CreateCartItem(ctx, &newItem); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &newItem, nil
}

// UpdateItem updates a cart line owned by the user.
func (c *Commands) UpdateItem(ctx context.Context, userID, cartItemID uint, in UpdateItemInput) error {
	cartItem, err := c.reader.FindCartItemForUser(ctx, userID, cartItemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("cart item not found")
		}
		return utils.ErrInternal(err)
	}

	if in.Quantity > 0 {
		product, err := c.reader.GetProductByID(ctx, cartItem.ProductID)
		if err != nil {
			return utils.ErrInternal(err)
		}
		if !ProductStockAvailable(*product, in.Quantity) {
			return utils.ErrBadRequest("insufficient stock")
		}
		cartItem.Quantity = in.Quantity
	}

	if in.Color != "" || in.Size != "" {
		cartItem.Color = in.Color
		cartItem.Size = in.Size
	}

	if err := c.writer.SaveCartItem(ctx, cartItem); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// RemoveItem deletes a cart line from the user's active cart.
func (c *Commands) RemoveItem(ctx context.Context, userID, cartItemID uint) error {
	rows, err := c.writer.RemoveCartItemForUser(ctx, userID, cartItemID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("cart item not found")
	}
	return nil
}

// Clear removes all items from the user's active cart.
func (c *Commands) Clear(ctx context.Context, userID uint) error {
	if err := c.writer.ClearActiveCartItems(ctx, userID); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// CleanAbandoned marks stale active carts as abandoned.
func (c *Commands) CleanAbandoned(ctx context.Context) error {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	return c.writer.MarkAbandonedCarts(ctx, cutoff)
}
