package i18n

import "testing"

func TestNormalizeSearchQuery_PersianDigitsAndYehKaf(t *testing.T) {
	got := NormalizeSearchQuery("آیفون ۱۲")
	if got != "آیفون 12" {
		t.Fatalf("got %q, want %q", got, "آیفون 12")
	}

	got = NormalizeSearchQuery("كتاب")
	if got != "کتاب" {
		t.Fatalf("got %q, want %q", got, "کتاب")
	}
}

func TestBuildSearchDocument_Deduplicates(t *testing.T) {
	doc := BuildSearchDocument("Watch", "Watch", "  Watch  ", "Chronograph")
	if doc != "Watch Chronograph" {
		t.Fatalf("got %q", doc)
	}
}
