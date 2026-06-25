package catalog

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// PrepareCreate validates domain rules and maps the DTO to a product model.
func (c *Commands) PrepareCreate(req dto.CreateProductRequest, slug string) (*models.Product, error) {
	if err := c.ValidateCreate(CreateProductInputFromDTO(req)); err != nil {
		return nil, err
	}
	return BuildCreateModel(req, slug), nil
}

// PersistCreate inserts a prepared product model.
func (c *Commands) PersistCreate(ctx context.Context, product *models.Product) error {
	return c.writer.CreateModel(ctx, product)
}

// Save persists product field changes.
func (c *Commands) Save(ctx context.Context, product *models.Product) error {
	return c.writer.SaveModel(ctx, product)
}

// UpdateStockColumn sets stock directly on the product row.
func (c *Commands) UpdateStockColumn(ctx context.Context, productID uint, stock int) error {
	return c.writer.UpdateStockColumn(ctx, productID, stock)
}

// ReplaceAttributes replaces all attributes for a product.
func (c *Commands) ReplaceAttributes(ctx context.Context, productID uint, inputs []dto.ProductAttributeInput) error {
	attrs := BuildAttributes(inputs)
	for i := range attrs {
		attrs[i].ProductID = productID
	}
	return c.writer.ReplaceAttributes(ctx, productID, attrs)
}

// BulkCreate inserts multiple product models in one transaction.
func (c *Commands) BulkCreate(ctx context.Context, products []*models.Product) error {
	return c.writer.BulkCreateModels(ctx, products)
}

// BulkDelete removes multiple products by id.
func (c *Commands) BulkDelete(ctx context.Context, ids []uint) (int64, error) {
	return c.writer.BulkDeleteByIDs(ctx, ids)
}

// Delete removes a single product by id.
func (c *Commands) Delete(ctx context.Context, id uint) (int64, error) {
	return c.writer.DeleteByID(ctx, id)
}
