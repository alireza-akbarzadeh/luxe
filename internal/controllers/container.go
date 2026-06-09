// Package controllers contains all the controller definitions for handling HTTP requests. Each controller corresponds to a specific domain (e.g., Auth, Product, Order) and contains methods for processing incoming requests, validating input, calling service layer functions, and returning appropriate HTTP responses. The Container struct aggregates all controllers for easy dependency injection into route setup.
package controllers

import (
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
	Wallet   *WalletController
	Payment  *PaymentProviderController
	Store    *StoreController
	Search   *SearchController
	Compare  *CompareController
	NavMenu  *NavMenuController
	Brand    *BrandController
	Settings *SettingController
}

// NewContainer initializes all controllers with their dependencies.
func NewContainer(db *gorm.DB, svc *services.Services) *Container {
	return &Container{
		Health:   NewHealthController(db),
		Search:   NewSearchController(svc.Search),
		Auth:     NewAuthController(svc.Auth),
		User:     NewUserController(svc.User, svc.Address),
		Cart:     NewCartController(svc.Cart),
		Product:  NewProductController(svc.Product, svc.UserLike),
		Compare:  NewCompareController(svc.Compare),
		Category: NewCategoryController(svc.Category),
		Order:    NewOrderController(svc.Order, svc.Checkout),
		Shipment: NewShipmentController(svc.Shipment),
		Page:     NewPageController(),
		Store:    NewStoreController(svc.Store, svc.Product),
		Account:  NewAccountController(svc.Address, svc.UserLike, svc.Order, svc.User),
		Coupon:   NewCouponController(svc.Coupon),
		Address:  NewAddressController(svc.Address),
		Menu:     NewMenuController(svc.Menu),
		Review:   NewReviewController(svc.Review),
		UserLike: NewUserLikeController(svc.UserLike, svc.Product),
		Wallet:   NewWallerController(svc.Wallet),
		Payment:  NewPaymentMethodController(svc.Payment),
		NavMenu:  NewNavMenuController(svc.NavMenu),
		Brand:    NewBrandController(svc.Brand),
		Settings: NewSettingController(svc.Settings),
	}
}
