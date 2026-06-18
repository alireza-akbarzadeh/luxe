package middleware

import "testing"

func TestPermissionDeniedMessage(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"orders.read", "You do not have read access for orders"},
		{"orders.write", "You do not have write access for orders"},
		{"products.read", "You do not have read access for products"},
		{"menus.write", "You do not have write access for menus"},
	}

	for _, tt := range tests {
		if got := PermissionDeniedMessage(tt.key); got != tt.want {
			t.Fatalf("PermissionDeniedMessage(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}
