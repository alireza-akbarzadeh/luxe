package integration

import (
	"testing"

	domaincart "github.com/alireza-akbarzadeh/luxe/internal/domain/cart"
	domaincheckout "github.com/alireza-akbarzadeh/luxe/internal/domain/checkout"
	domainorder "github.com/alireza-akbarzadeh/luxe/internal/domain/order"
)

// TestCommerceDomainRules verifies shared commerce domain invariants used by checkout.
func TestCommerceDomainRules(t *testing.T) {
	cartSvc := domaincart.NewService()
	if err := cartSvc.ValidateCheckout(domaincart.Cart{Items: []domaincart.Item{{Quantity: 1, UnitCents: 500}}}); err != nil {
		t.Fatalf("cart: %v", err)
	}

	checkoutSvc := domaincheckout.NewService()
	if err := checkoutSvc.Validate(domaincheckout.CheckoutInput{
		UserID: 1, CartTotalCents: 500, Currency: "USD", PaymentMethod: "wallet",
	}); err != nil {
		t.Fatalf("checkout: %v", err)
	}

	orderSvc := domainorder.NewService()
	if err := orderSvc.CanCancel(domainorder.Order{Status: "pending"}); err != nil {
		t.Fatalf("order cancel pending: %v", err)
	}
	if err := orderSvc.CanCancel(domainorder.Order{Status: "delivered"}); err == nil {
		t.Fatal("expected delivered order to be non-cancellable")
	}
}
