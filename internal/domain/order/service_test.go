package order

import "testing"

func TestCanCancel(t *testing.T) {
	svc := NewService()
	if err := svc.CanCancel(Order{Status: "delivered"}); err != ErrCannotCancel {
		t.Fatalf("got %v", err)
	}
	if err := svc.CanCancel(Order{Status: "pending"}); err != nil {
		t.Fatalf("unexpected %v", err)
	}
}
