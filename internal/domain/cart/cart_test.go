package cart

import "testing"

func TestValidateCheckout(t *testing.T) {
	svc := NewService()
	if err := svc.ValidateCheckout(Cart{}); err != ErrEmptyCart {
		t.Fatalf("got %v", err)
	}
	if err := svc.ValidateCheckout(Cart{Items: []Item{{ProductID: 1, Quantity: 1, UnitCents: 100}}}); err != nil {
		t.Fatalf("unexpected %v", err)
	}
}
