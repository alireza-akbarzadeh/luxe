package importdata_test

import (
	"bytes"
	"testing"

	importdata "github.com/alireza-akbarzadeh/luxe/internal/application/import"
	"github.com/xuri/excelize/v2"
)

func TestProductTemplateHasExactImportHeaders(t *testing.T) {
	svc := importdata.NewService(nil, nil)
	data, err := svc.ProductTemplate()
	if err != nil {
		t.Fatalf("ProductTemplate: %v", err)
	}
	if len(data) < 100 {
		t.Fatalf("template too small (%d bytes)", len(data))
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("open template: %v", err)
	}
	defer f.Close()

	wantHeaders := []string{
		"name", "sku", "price", "stock", "description",
		"status", "category_id", "brand_id", "store_id",
		"compare_at_price", "barcode", "low_stock_threshold", "weight",
	}

	rows, err := f.GetRows("Products")
	if err != nil {
		t.Fatalf("Products sheet: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("expected header + sample rows, got %d", len(rows))
	}
	if len(rows[0]) < len(wantHeaders) {
		t.Fatalf("header too short: %v", rows[0])
	}
	for i, want := range wantHeaders {
		if rows[0][i] != want {
			t.Fatalf("header[%d]=%q, want %q", i, rows[0][i], want)
		}
	}

	guide, err := f.GetRows("Column guide")
	if err != nil {
		t.Fatalf("Column guide sheet: %v", err)
	}
	if len(guide) < len(wantHeaders)+1 {
		t.Fatalf("column guide incomplete: %d rows", len(guide))
	}
}
