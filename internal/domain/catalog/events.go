package catalog

// ProductPublished is emitted when a product becomes sellable.
type ProductPublished struct {
	ProductID uint
	Slug      string
}

// ProductPriceChanged is emitted when price changes.
type ProductPriceChanged struct {
	ProductID  uint
	PriceCents int64
	Currency   string
}
