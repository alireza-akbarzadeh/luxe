package catalog

import "testing"

func TestServiceValidateCreate(t *testing.T) {
	svc := NewService()
	tests := []struct {
		name    string
		in      CreateProductInput
		wantErr error
	}{
		{
			name: "ok",
			in: CreateProductInput{
				Name: "Watch", PriceCents: 1999, Currency: "USD", SKU: "W-1",
			},
		},
		{name: "empty name", in: CreateProductInput{PriceCents: 100, SKU: "x"}, wantErr: ErrInvalidProductName},
		{name: "zero price", in: CreateProductInput{Name: "x", SKU: "x"}, wantErr: ErrInvalidPrice},
		{name: "empty sku", in: CreateProductInput{Name: "x", PriceCents: 100}, wantErr: ErrInvalidSKU},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateCreate(tt.in)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
		})
	}
}
