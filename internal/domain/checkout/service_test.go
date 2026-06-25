package checkout

import "testing"

func TestValidate(t *testing.T) {
	svc := NewService()
	err := svc.Validate(CheckoutInput{UserID: 1, CartTotalCents: 100, PaymentMethod: "stripe"})
	if err != nil {
		t.Fatal(err)
	}
	if svc.Validate(CheckoutInput{}) == nil {
		t.Fatal("expected error")
	}
}
