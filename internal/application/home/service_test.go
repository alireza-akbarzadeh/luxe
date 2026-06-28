package home

import (
	"context"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
)

func TestMergeUniqueCategoryIDs(t *testing.T) {
	t.Parallel()

	got := mergeUniqueCategoryIDs([]uint{1, 2}, []uint{2, 3, 0}, []uint{4, 1})
	want := []uint{1, 2, 3, 4}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestCapCategoryItems(t *testing.T) {
	t.Parallel()

	items := []dto.HomeCategoryItem{{ID: 1}, {ID: 2}, {ID: 3}}
	got := capCategoryItems(items, 2)
	if len(got) != 2 || got[0].ID != 1 || got[1].ID != 2 {
		t.Fatalf("unexpected cap result: %+v", got)
	}
}

func TestBuildForYouCategoriesGuestUsesPopular(t *testing.T) {
	t.Parallel()

	svc := &Service{cache: newTTLCache()}
	popular := []dto.HomeCategoryItem{
		{ID: 10, Name: "Women", Slug: "women"},
		{ID: 11, Name: "Men", Slug: "men"},
	}

	got, err := svc.buildForYouCategories(context.Background(), nil, 8, popular)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0].ID != 10 || got[1].ID != 11 {
		t.Fatalf("got %+v, want popular copy", got)
	}
}
