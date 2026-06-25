package services

import (
	"context"

	appcart "github.com/alireza-akbarzadeh/luxe/internal/application/cart"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type AddItemRequest struct {
	ProductID uint `json:"product_id" validate:"required,gt=0"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`

	Color string `json:"color"`
	Size  string `json:"size"`
}

type UpdateCartItemRequest struct {
	Quantity int    `json:"quantity" validate:"omitempty,gt=0"`
	Color    string `json:"color"`
	Size     string `json:"size"`
}

type CartServiceInterface interface {
	GetOrCreateCart(ctx context.Context, userID uint) (*models.Cart, error)
	AddItem(ctx context.Context, userID uint, req AddItemRequest) (*models.CartItem, error)
	UpdateCartItem(ctx context.Context, userID uint, cartItemID uint, req UpdateCartItemRequest) error
	RemoveItem(ctx context.Context, userID uint, cartItemID uint) error
	GetCart(ctx context.Context, userID uint) (*models.Cart, error)
	ClearCart(ctx context.Context, userID uint) error
	CleanAbandonedCarts(ctx context.Context) error
}

type cartService struct {
	commands *appcart.Commands
	queries  *appcart.Queries
}

func NewCartService(db *gorm.DB) CartServiceInterface {
	repo := postgres.NewCartRepository(db)
	return &cartService{
		commands: appcart.NewCommands(repo, repo),
		queries:  appcart.NewQueries(repo),
	}
}

func (s *cartService) GetOrCreateCart(ctx context.Context, userID uint) (*models.Cart, error) {
	return s.commands.GetOrCreateCart(ctx, userID)
}

func (s *cartService) AddItem(ctx context.Context, userID uint, req AddItemRequest) (*models.CartItem, error) {
	return s.commands.AddItem(ctx, userID, appcart.AddItemInput{
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	})
}

func (s *cartService) UpdateCartItem(ctx context.Context, userID uint, cartItemID uint, req UpdateCartItemRequest) error {
	return s.commands.UpdateItem(ctx, userID, cartItemID, appcart.UpdateItemInput{
		Quantity: req.Quantity,
		Color:    req.Color,
		Size:     req.Size,
	})
}

func (s *cartService) RemoveItem(ctx context.Context, userID uint, cartItemID uint) error {
	return s.commands.RemoveItem(ctx, userID, cartItemID)
}

func (s *cartService) GetCart(ctx context.Context, userID uint) (*models.Cart, error) {
	return s.queries.Get(ctx, userID)
}

func (s *cartService) ClearCart(ctx context.Context, userID uint) error {
	return s.commands.Clear(ctx, userID)
}

func (s *cartService) CleanAbandonedCarts(ctx context.Context) error {
	return s.commands.CleanAbandoned(ctx)
}
