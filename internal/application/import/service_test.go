package importdata

import (
	"bytes"
	"errors"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/xuri/excelize/v2"
)

type mockProductCreator struct {
	createFn func(req dto.CreateProductRequest) (*models.Product, error)
}

func (m *mockProductCreator) Create(req dto.CreateProductRequest) (*models.Product, error) {
	return m.createFn(req)
}

type mockCategoryCreator struct {
	createFn func(req dto.CreateCategoryRequest) (*models.Category, error)
}

func (m *mockCategoryCreator) Create(req dto.CreateCategoryRequest) (*models.Category, error) {
	return m.createFn(req)
}

func makeExcel(sheet string, rows [][]string) *bytes.Buffer {
	f := excelize.NewFile()
	_ = f.SetSheetName("Sheet1", sheet)
	for ri, row := range rows {
		for ci, val := range row {
			cell, _ := excelize.CoordinatesToCellName(ci+1, ri+1)
			_ = f.SetCellValue(sheet, cell, val)
		}
	}
	buf, _ := f.WriteToBuffer()
	return buf
}

func TestImportCategoriesFromExcel_AllCreated(t *testing.T) {
	var created []string
	svc := NewService(
		&mockProductCreator{},
		&mockCategoryCreator{
			createFn: func(req dto.CreateCategoryRequest) (*models.Category, error) {
				created = append(created, req.Name)
				id := uint(len(created))
				return &models.Category{ID: id, Name: req.Name}, nil
			},
		},
	)

	buf := makeExcel("Categories", [][]string{
		{"name", "slug", "description", "parent_id", "is_active"},
		{"Clothing", "clothing", "All clothing", "", "true"},
		{"Shoes", "shoes", "Footwear", "", "true"},
	})

	summary, err := svc.ImportCategoriesFromExcel(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalRows != 2 {
		t.Errorf("TotalRows: want 2, got %d", summary.TotalRows)
	}
	if summary.Created != 2 {
		t.Errorf("Created: want 2, got %d", summary.Created)
	}
	if summary.Failed != 0 || summary.Skipped != 0 {
		t.Errorf("want 0 failed/skipped, got failed=%d skipped=%d", summary.Failed, summary.Skipped)
	}
}

func TestImportCategoriesFromExcel_EmptyNameSkipped(t *testing.T) {
	svc := NewService(&mockProductCreator{}, &mockCategoryCreator{
		createFn: func(req dto.CreateCategoryRequest) (*models.Category, error) {
			return &models.Category{ID: 1, Name: req.Name}, nil
		},
	})

	buf := makeExcel("Categories", [][]string{
		{"name", "slug", "description", "parent_id", "is_active"},
		{"", "clothing", "", "", ""},
		{"Clothing", "clothing", "", "", "true"},
	})

	summary, err := svc.ImportCategoriesFromExcel(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Skipped != 1 {
		t.Errorf("Skipped: want 1, got %d", summary.Skipped)
	}
	if summary.Created != 1 {
		t.Errorf("Created: want 1, got %d", summary.Created)
	}
}

func TestImportCategoriesFromExcel_ServiceError(t *testing.T) {
	svc := NewService(&mockProductCreator{}, &mockCategoryCreator{
		createFn: func(_ dto.CreateCategoryRequest) (*models.Category, error) {
			return nil, errors.New("db error")
		},
	})

	buf := makeExcel("Categories", [][]string{
		{"name", "slug", "description", "parent_id", "is_active"},
		{"Clothing", "clothing", "", "", "true"},
	})

	summary, err := svc.ImportCategoriesFromExcel(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Failed != 1 {
		t.Errorf("Failed: want 1, got %d", summary.Failed)
	}
	if summary.Rows[0].Error == "" {
		t.Error("expected error message in row result")
	}
}

func TestImportCategoriesFromExcel_HeaderOnly(t *testing.T) {
	svc := NewService(&mockProductCreator{}, &mockCategoryCreator{})

	buf := makeExcel("Categories", [][]string{
		{"name", "slug", "description", "parent_id", "is_active"},
	})

	_, err := svc.ImportCategoriesFromExcel(buf)
	if err == nil {
		t.Fatal("expected error for header-only file, got nil")
	}
	appErr, ok := err.(*utils.AppError)
	if !ok || appErr.Code != 400 {
		t.Errorf("expected 400 AppError, got %v", err)
	}
}

func TestImportProductsFromExcel_Created(t *testing.T) {
	svc := NewService(
		&mockProductCreator{
			createFn: func(req dto.CreateProductRequest) (*models.Product, error) {
				return &models.Product{ID: 1, Name: req.Name}, nil
			},
		},
		&mockCategoryCreator{},
	)

	buf := makeExcel("Products", [][]string{
		{"name", "sku", "price", "stock", "description", "status", "category_id", "brand_id", "store_id",
			"compare_at_price", "barcode", "low_stock_threshold", "weight"},
		{"Widget", "WG-001", "29.99", "10", "A widget", "active", "", "", "", "", "", "", ""},
	})

	summary, err := svc.ImportProductsFromExcel(buf, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Created != 1 {
		t.Errorf("Created: want 1, got %d", summary.Created)
	}
}

func TestImportProductsFromExcel_InvalidPrice(t *testing.T) {
	svc := NewService(&mockProductCreator{}, &mockCategoryCreator{})

	buf := makeExcel("Products", [][]string{
		{"name", "sku", "price", "stock", "description", "status", "category_id", "brand_id", "store_id",
			"compare_at_price", "barcode", "low_stock_threshold", "weight"},
		{"Widget", "WG-001", "not-a-number", "10", "", "active", "", "", "", "", "", "", ""},
	})

	summary, err := svc.ImportProductsFromExcel(buf, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Failed != 1 {
		t.Errorf("Failed: want 1, got %d", summary.Failed)
	}
}

func TestImportProductsFromExcel_StoreIDFallback(t *testing.T) {
	var gotStoreID *uint
	svc := NewService(
		&mockProductCreator{
			createFn: func(req dto.CreateProductRequest) (*models.Product, error) {
				gotStoreID = req.StoreID
				return &models.Product{ID: 1, Name: req.Name}, nil
			},
		},
		&mockCategoryCreator{},
	)

	buf := makeExcel("Products", [][]string{
		{"name", "sku", "price", "stock", "description", "status", "category_id", "brand_id", "store_id",
			"compare_at_price", "barcode", "low_stock_threshold", "weight"},
		{"Widget", "WG-001", "9.99", "5", "", "active", "", "", "", "", "", "", ""},
	})

	_, err := svc.ImportProductsFromExcel(buf, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotStoreID == nil || *gotStoreID != 42 {
		t.Errorf("expected store_id fallback 42, got %v", gotStoreID)
	}
}

func TestProductTemplate_ReturnsValidXLSX(t *testing.T) {
	svc := NewService(&mockProductCreator{}, &mockCategoryCreator{})
	data, err := svc.ProductTemplate()
	if err != nil {
		t.Fatalf("ProductTemplate error: %v", err)
	}
	if len(data) == 0 {
		t.Error("ProductTemplate returned empty bytes")
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ProductTemplate produced invalid xlsx: %v", err)
	}
	defer f.Close()
}

func TestCategoryTemplate_ReturnsValidXLSX(t *testing.T) {
	svc := NewService(&mockProductCreator{}, &mockCategoryCreator{})
	data, err := svc.CategoryTemplate()
	if err != nil {
		t.Fatalf("CategoryTemplate error: %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("CategoryTemplate produced invalid xlsx: %v", err)
	}
	defer f.Close()
}
