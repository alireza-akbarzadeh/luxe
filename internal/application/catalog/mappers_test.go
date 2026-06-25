package catalog_test

import (
	"context"
	"errors"
	"testing"

	appcatalog "github.com/alireza-akbarzadeh/luxe/internal/application/catalog"
	domaincatalog "github.com/alireza-akbarzadeh/luxe/internal/domain/catalog"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

func TestBuildCreateModel_Defaults(t *testing.T) {
	req := dto.CreateProductRequest{
		Name:  "Luxury Watch",
		Price: 199.99,
		SKU:   "W-100",
		Stock: 10,
	}
	product := appcatalog.BuildCreateModel(req, "luxury-watch")

	if product.Slug != "luxury-watch" {
		t.Fatalf("slug = %q", product.Slug)
	}
	if product.Status != "draft" {
		t.Fatalf("status = %q, want draft", product.Status)
	}
	if product.LowStockThreshold != 5 {
		t.Fatalf("low stock threshold = %d, want 5", product.LowStockThreshold)
	}
	if product.IsNew {
		t.Fatal("expected is_new false by default")
	}
}

func TestBuildAttributes(t *testing.T) {
	attrs := appcatalog.BuildAttributes([]dto.ProductAttributeInput{
		{Name: "Color", Values: []string{"Gold", "Silver"}},
	})
	if len(attrs) != 1 || attrs[0].Name != "Color" || len(attrs[0].Values) != 2 {
		t.Fatalf("unexpected attrs: %+v", attrs)
	}
}

func TestApplyUpdateDTO(t *testing.T) {
	name := "Updated Name"
	price := 49.99
	product := &models.Product{Name: "Old", Slug: "old", Price: 10}
	req := dto.UpdateProductRequest{Name: &name, Price: &price}

	appcatalog.ApplyUpdateDTO(product, req, "updated-name")

	if product.Name != name || product.Slug != "updated-name" || product.Price != price {
		t.Fatalf("product not updated: %+v", product)
	}
}

func TestCreateProductInputFromDTO(t *testing.T) {
	in := appcatalog.CreateProductInputFromDTO(dto.CreateProductRequest{
		Name: "Item", Price: 12.5, SKU: "SKU-1", Stock: 3,
	})
	if in.Name != "Item" || in.PriceCents != 1250 || in.SKU != "SKU-1" {
		t.Fatalf("unexpected input: %+v", in)
	}
}

type mockWriter struct {
	created []*models.Product
	deleted int64
}

func (m *mockWriter) CreateModel(_ context.Context, product *models.Product) error {
	m.created = append(m.created, product)
	product.ID = uint(len(m.created))
	return nil
}
func (m *mockWriter) SaveModel(context.Context, *models.Product) error             { return nil }
func (m *mockWriter) UpdateStockColumn(context.Context, uint, int) error           { return nil }
func (m *mockWriter) ReplaceAttributes(context.Context, uint, []models.ProductAttribute) error {
	return nil
}
func (m *mockWriter) BulkCreateModels(_ context.Context, products []*models.Product) error {
	m.created = append(m.created, products...)
	return nil
}
func (m *mockWriter) BulkDeleteByIDs(_ context.Context, _ []uint) (int64, error) {
	m.deleted = 2
	return 2, nil
}
func (m *mockWriter) DeleteByID(_ context.Context, _ uint) (int64, error) { return 1, nil }

func TestCommandsPrepareAndPersistCreate(t *testing.T) {
	writer := &mockWriter{}
	cmds := appcatalog.NewCommands(domaincatalog.NewService(), nil, writer)

	product, err := cmds.PrepareCreate(dto.CreateProductRequest{
		Name: "Valid", Price: 10, SKU: "X-1", Stock: 1,
	}, "valid")
	if err != nil {
		t.Fatalf("PrepareCreate: %v", err)
	}
	if err := cmds.PersistCreate(context.Background(), product); err != nil {
		t.Fatalf("PersistCreate: %v", err)
	}
	if len(writer.created) != 1 {
		t.Fatalf("created count = %d", len(writer.created))
	}
}

func TestCommandsPrepareCreateValidation(t *testing.T) {
	cmds := appcatalog.NewCommands(domaincatalog.NewService(), nil, &mockWriter{})
	_, err := cmds.PrepareCreate(dto.CreateProductRequest{Name: "", Price: 10, SKU: "x"}, "x")
	if !errors.Is(err, domaincatalog.ErrInvalidProductName) {
		t.Fatalf("err = %v", err)
	}
}

func TestCommandsBulkDelete(t *testing.T) {
	writer := &mockWriter{}
	cmds := appcatalog.NewCommands(domaincatalog.NewService(), nil, writer)
	rows, err := cmds.BulkDelete(context.Background(), []uint{1, 2})
	if err != nil || rows != 2 {
		t.Fatalf("BulkDelete rows=%d err=%v", rows, err)
	}
}
