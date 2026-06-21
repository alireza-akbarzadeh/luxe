package i18n_test

import (
	"context"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
)

func TestParseAcceptLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		header string
		want   string
	}{
		{"", "en"},
		{"en", "en"},
		{"es-ES,es;q=0.9,en;q=0.8", "es"},
		{"fa-IR,fa;q=0.9", "fa"},
		{"de-DE,de;q=0.9", "en"},
	}

	for _, tt := range tests {
		if got := i18n.ParseAcceptLanguage(tt.header); got != tt.want {
			t.Errorf("ParseAcceptLanguage(%q) = %q, want %q", tt.header, got, tt.want)
		}
	}
}

func TestTranslate(t *testing.T) {
	if err := i18n.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	ctx := i18n.WithLocale(context.Background(), "es")
	got := i18n.Translate(ctx, constants.ErrNotFound)
	if got != "recurso no encontrado" {
		t.Fatalf("Translate(es) = %q", got)
	}

	ctxFa := i18n.WithLocale(context.Background(), "fa")
	gotFa := i18n.Translate(ctxFa, constants.ErrInvalidCredentials)
	if gotFa != "ایمیل یا رمز عبور نامعتبر است" {
		t.Fatalf("Translate(fa) = %q", gotFa)
	}

	unknown := i18n.Translate(ctx, "custom service message")
	if unknown != "custom service message" {
		t.Fatalf("unknown message should pass through, got %q", unknown)
	}
}

func TestT(t *testing.T) {
	if err := i18n.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	ctx := i18n.WithLocale(context.Background(), "fa")
	if got := i18n.T(ctx, "success"); got != "موفقیت" {
		t.Fatalf("T(fa, success) = %q", got)
	}
}
