package i18n

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalizedMapResolve(t *testing.T) {
	m := LocalizedMap{
		"en": "Shop",
		"fa": "فروشگاه",
	}

	ctx := WithLocale(context.Background(), "fa")
	assert.Equal(t, "فروشگاه", m.Resolve(ctx, "Shop"))

	ctxEs := WithLocale(context.Background(), "es")
	assert.Equal(t, "Shop", m.Resolve(ctxEs, "Shop"))
}

func TestMergeLocalizedPreservesExistingLocales(t *testing.T) {
	existing := LocalizedMap{"en": "Shop", "fa": "فروشگاه"}
	incoming := LocalizedMap{"es": "Tienda"}

	merged := MergeLocalized(existing, incoming, "Shop")
	assert.Equal(t, "Shop", merged["en"])
	assert.Equal(t, "فروشگاه", merged["fa"])
	assert.Equal(t, "Tienda", merged["es"])
}

func TestParseLocalizedFieldString(t *testing.T) {
	m, fallback := ParseLocalizedField([]byte(`"Women"`))
	assert.Equal(t, "Women", fallback)
	assert.Equal(t, "Women", m["en"])
}

func TestParseLocalizedFieldMap(t *testing.T) {
	m, fallback := ParseLocalizedField([]byte(`{"en":"Women","fa":"زنانه"}`))
	assert.Equal(t, "Women", fallback)
	assert.Equal(t, "زنانه", m["fa"])
}
