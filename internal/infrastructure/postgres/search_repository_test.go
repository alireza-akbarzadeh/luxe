package postgres

import (
	"strings"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
)

func TestProductSearchWhere_HeritageMatchesNameFallback(t *testing.T) {
	clause, args := productSearchWhere("Heritage")

	if clause == "1=0" {
		t.Fatal("expected searchable clause for Heritage")
	}
	if !strings.Contains(clause, "products.name ILIKE") {
		t.Fatalf("expected name ILIKE fallback in clause: %s", clause)
	}
	if len(args) < 2 {
		t.Fatalf("expected query args, got %d", len(args))
	}
	if got := i18n.NormalizeSearchQuery("Heritage"); got != "Heritage" {
		t.Fatalf("normalize query = %q", got)
	}
	like := "%" + i18n.NormalizeSearchQuery("Heritage") + "%"
	foundLike := false
	for _, arg := range args {
		if s, ok := arg.(string); ok && s == like {
			foundLike = true
			break
		}
	}
	if !foundLike {
		t.Fatalf("expected ILIKE pattern %q in args %v", like, args)
	}
}

func TestProductSearchWhere_EmptyQuery(t *testing.T) {
	clause, args := productSearchWhere("   ")
	if clause != "1=0" || args != nil {
		t.Fatalf("empty query should be 1=0, got clause=%q args=%v", clause, args)
	}
}
