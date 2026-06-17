package services

import (
	"bytes"
	"errors"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/xuri/excelize/v2"
)

// ─── Minimal mocks ───────────────────────────────────────────────────────────

type mockProductSvc struct {
	createFn func(req dto.CreateProductRequest) (*models.Product, error)
}

func (m *mockProductSvc) Create(req dto.CreateProductRequest) (*models.Product, error) {
	return m.createFn(req)
}
func (m *mockProductSvc) List(_ int, _ int, _ dto.ProductListFilters) ([]*models.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockProductSvc) BulkCreate(_ []dto.CreateProductRequest) ([]*models.Product, error) {
	return nil, nil
}
func (m *mockProductSvc) GetByID(_ uint) (*models.Product, error)   { return nil, nil }
func (m *mockProductSvc) GetBySlug(_ string) (*models.Product, error) { return nil, nil }
func (m *mockProductSvc) Update(_ uint, _ dto.UpdateProductRequest) (*models.Product, error) {
	return nil, nil
}
func (m *mockProductSvc) Delete(_ uint) error                                            { return nil }
func (m *mockProductSvc) BulkDelete(_ []uint) error                                      { return nil }
func (m *mockProductSvc) CheckLowStockAndAlert() error                                   { return nil }
func (m *mockProductSvc) GetRelated(_ uint, _ int) ([]*models.Product, error)             { return nil, nil }
func (m *mockProductSvc) GetSuggestions(_ []uint, _ int) ([]*models.Product, error)       { return nil, nil }
func (m *mockProductSvc) GetByStoreID(_ uint, _ int, _ int, _ dto.ProductListFilters) ([]*models.Product, int64, error) {
	return nil, 0, nil
}

type mockCategorySvc struct {
	createFn func(req dto.CreateCategoryRequest) (*models.Category, error)
}

func (m *mockCategorySvc) Create(req dto.CreateCategoryRequest) (*models.Category, error) {
	return m.createFn(req)
}
func (m *mockCategorySvc) GetByID(_ uint) (*models.Category, error)    { return nil, nil }
func (m *mockCategorySvc) GetBySlug(_ string) (*models.Category, error) { return nil, nil }
func (m *mockCategorySvc) Update(_ uint, _ dto.UpdateCategoryRequest) (*models.Category, error) {
	return nil, nil
}
func (m *mockCategorySvc) Delete(_ uint) error                                              { return nil }
func (m *mockCategorySvc) List(_ dto.CategoryListFilters) ([]models.Category, int64, error) { return nil, 0, nil }
func (m *mockCategorySvc) BulkCreate(_ []dto.CreateCategoryRequest) ([]*models.Category, error) {
	return nil, nil
}
func (m *mockCategorySvc) BulkDelete(_ []uint) error { return nil }

// ─── Helpers ─────────────────────────────────────────────────────────────────

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

func ptrUint(v uint) *uint { return &v }

// ─── ImportCategoriesFromExcel ────────────────────────────────────────────────

func TestImportCategoriesFromExcel_AllCreated(t *testing.T) {
	var created []string
	svc := NewImportService(
		&mockProductSvc{},
		&mockCategorySvc{
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
	svc := NewImportService(&mockProductSvc{}, &mockCategorySvc{
		createFn: func(req dto.CreateCategoryRequest) (*models.Category, error) {
			return &models.Category{ID: 1, Name: req.Name}, nil
		},
	})

	buf := makeExcel("Categories", [][]string{
		{"name", "slug", "description", "parent_id", "is_active"},
		{"",        "",   "",           "",          ""},         // empty → skipped
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
	svc := NewImportService(&mockProductSvc{}, &mockCategorySvc{
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
	svc := NewImportService(&mockProductSvc{}, &mockCategorySvc{})

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

// ─── ImportProductsFromExcel ──────────────────────────────────────────────────

func TestImportProductsFromExcel_Created(t *testing.T) {
	svc := NewImportService(
		&mockProductSvc{
			createFn: func(req dto.CreateProductRequest) (*models.Product, error) {
				return &models.Product{ID: 1, Name: req.Name}, nil
			},
		},
		&mockCategorySvc{},
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
	svc := NewImportService(&mockProductSvc{}, &mockCategorySvc{})

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
	svc := NewImportService(
		&mockProductSvc{
			createFn: func(req dto.CreateProductRequest) (*models.Product, error) {
				gotStoreID = req.StoreID
				return &models.Product{ID: 1, Name: req.Name}, nil
			},
		},
		&mockCategorySvc{},
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

// ─── Templates ───────────────────────────────────────────────────────────────

func TestProductTemplate_ReturnsValidXLSX(t *testing.T) {
	svc := NewImportService(&mockProductSvc{}, &mockCategorySvc{})
	data, err := svc.ProductTemplate()
	if err != nil {
		t.Fatalf("ProductTemplate error: %v", err)
	}
	if len(data) == 0 {
		t.Error("ProductTemplate returned empty bytes")
	}
	// Verify it is a valid Excel file by re-opening it.
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ProductTemplate produced invalid xlsx: %v", err)
	}
	defer f.Close()
}

func TestCategoryTemplate_ReturnsValidXLSX(t *testing.T) {
	svc := NewImportService(&mockProductSvc{}, &mockCategorySvc{})
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
