// Package controllers contains all the controller definitions for handling HTTP requests. Each controller corresponds to a specific domain (e.g., Auth, Product, Order) and contains methods for processing incoming requests, validating input, calling service layer functions, and returning appropriate HTTP responses. The Container struct aggregates all controllers for easy dependency injection into route setup.
package controllers

import (
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"gorm.io/gorm"
)

type Container struct {
	Health   *HealthController
	Auth     *AuthController
	User     *UserController
	Page     *PageController
	Cart     *CartController
	Product  *ProductController
	Category *CategoryController
	Order    *OrderController
	Shipment *ShipmentController
	Coupon   *CouponController
	Address  *AddressController
	Menu     *MenuController
	Review   *ReviewController
	UserLike *UserLikeController
	Account  *AccountController
	Wallet   *WalletCotroller
}

// NewContainer initializes all controllers with their dependencies.
func NewContainer(db *gorm.DB, cfg *config.Config, svc *services.Services) *Container {
	return &Container{
		Health:   NewHealthController(db),
		Auth:     NewAuthController(svc.Auth),
		User:     NewUserController(svc.User, svc.Address),
		Cart:     NewCartController(svc.Cart),
		Product:  NewProductController(svc.Product, svc.UserLike),
		Category: NewCategoryController(svc.Category),
		Order:    NewOrderController(svc.Order),
		Shipment: NewShipmentController(svc.Shipment),
		Page:     NewPageController(),
		Account:  NewAccountController(svc.Address, svc.UserLike, svc.Order, svc.User),
		Coupon:   NewCouponController(svc.Coupon),
		Address:  NewAddressController(svc.Address),
		Menu:     NewMenuController(svc.Menu),
		Review:   NewReviewController(svc.Review),
		UserLike: NewUserLikeController(svc.UserLike, svc.Product),
		Wallet:   NewWallerController(svc.Wallet),
	}
}
