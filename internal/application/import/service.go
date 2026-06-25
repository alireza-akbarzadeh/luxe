package importdata

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/go-playground/validator/v10"
	"github.com/xuri/excelize/v2"
)

// ProductCreator creates products during bulk import.
type ProductCreator interface {
	Create(req dto.CreateProductRequest) (*models.Product, error)
}

// CategoryCreator creates categories during bulk import.
type CategoryCreator interface {
	Create(req dto.CreateCategoryRequest) (*models.Category, error)
}

// Service handles Excel-based bulk import flows.
type Service struct {
	productSvc  ProductCreator
	categorySvc CategoryCreator
	validate    *validator.Validate
}

// NewService wires product and category creators for import.
func NewService(productSvc ProductCreator, categorySvc CategoryCreator) *Service {
	return &Service{
		productSvc:  productSvc,
		categorySvc: categorySvc,
		validate:    validator.New(),
	}
}

// productColumns defines the expected header row for product imports.
var productColumns = []string{
	"name", "sku", "price", "stock", "description",
	"status", "category_id", "brand_id", "store_id",
	"compare_at_price", "barcode", "low_stock_threshold", "weight",
}

func (s *Service) ImportProductsFromExcel(r io.Reader, storeID uint) (*dto.ImportSummary, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, utils.ErrBadRequest("invalid Excel file: " + err.Error())
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, utils.ErrBadRequest("cannot read sheet: " + err.Error())
	}
	if len(rows) < 2 {
		return nil, utils.ErrBadRequest("file has no data rows (expected header + at least one data row)")
	}

	summary := &dto.ImportSummary{
		TotalRows: len(rows) - 1, // exclude header
		Rows:      make([]dto.ImportRowResult, 0, len(rows)-1),
	}

	for i, row := range rows[1:] { // skip header row
		rowNum := i + 2 // Excel row number (1-based, header=1)
		cells := padCells(row, len(productColumns))

		name := strings.TrimSpace(cells[0])
		if name == "" {
			result := dto.ImportRowResult{Row: rowNum, Status: "skipped", Error: "empty name — row skipped"}
			summary.Rows = append(summary.Rows, result)
			summary.Skipped++
			continue
		}

		price, priceErr := strconv.ParseFloat(cells[2], 64)
		if priceErr != nil {
			result := dto.ImportRowResult{Row: rowNum, Status: "failed", Name: name,
				Error: fmt.Sprintf("invalid price %q: %v", cells[2], priceErr)}
			summary.Rows = append(summary.Rows, result)
			summary.Failed++
			continue
		}

		stock, _ := strconv.Atoi(cells[3])
		status := strings.TrimSpace(cells[5])
		if status == "" {
			status = constants.ProductStatusActive
		}

		req := dto.CreateProductRequest{
			Name:        name,
			SKU:         strings.TrimSpace(cells[1]),
			Price:       price,
			Stock:       stock,
			Description: strings.TrimSpace(cells[4]),
			Status:      status,
			Barcode:     strings.TrimSpace(cells[10]),
		}

		if v := strings.TrimSpace(cells[6]); v != "" {
			if id, err := strconv.ParseUint(v, 10, 64); err == nil {
				uid := uint(id)
				req.CategoryID = &uid
			}
		}
		if v := strings.TrimSpace(cells[7]); v != "" {
			if id, err := strconv.ParseUint(v, 10, 64); err == nil {
				uid := uint(id)
				req.BrandID = &uid
			}
		}
		// Use file-provided store_id if present, otherwise fall back to the one from the route.
		if v := strings.TrimSpace(cells[8]); v != "" {
			if id, err := strconv.ParseUint(v, 10, 64); err == nil {
				uid := uint(id)
				req.StoreID = &uid
			}
		} else if storeID > 0 {
			req.StoreID = &storeID
		}
		if v := strings.TrimSpace(cells[9]); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				req.CompareAtPrice = &f
			}
		}
		if v := strings.TrimSpace(cells[11]); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				req.LowStockThreshold = n
			}
		}
		if v := strings.TrimSpace(cells[12]); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				req.Weight = &f
			}
		}

		if err := s.validate.Struct(req); err != nil {
			result := dto.ImportRowResult{Row: rowNum, Status: "failed", Name: name, Error: err.Error()}
			summary.Rows = append(summary.Rows, result)
			summary.Failed++
			continue
		}

		created, err := s.productSvc.Create(req)
		if err != nil {
			result := dto.ImportRowResult{Row: rowNum, Status: "failed", Name: name, Error: serviceErrMsg(err)}
			summary.Rows = append(summary.Rows, result)
			summary.Failed++
			continue
		}

		result := dto.ImportRowResult{Row: rowNum, Status: "created", Name: name, ID: &created.ID}
		summary.Rows = append(summary.Rows, result)
		summary.Created++
	}
	return summary, nil
}

// ─── Category import ──────────────────────────────────────────────────────────

var categoryColumns = []string{"name", "slug", "description", "parent_id", "is_active"}

func (s *Service) ImportCategoriesFromExcel(r io.Reader) (*dto.ImportSummary, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, utils.ErrBadRequest("invalid Excel file: " + err.Error())
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, utils.ErrBadRequest("cannot read sheet: " + err.Error())
	}
	if len(rows) < 2 {
		return nil, utils.ErrBadRequest("file has no data rows (expected header + at least one data row)")
	}

	summary := &dto.ImportSummary{
		TotalRows: len(rows) - 1,
		Rows:      make([]dto.ImportRowResult, 0, len(rows)-1),
	}

	for i, row := range rows[1:] {
		rowNum := i + 2
		cells := padCells(row, len(categoryColumns))

		name := strings.TrimSpace(cells[0])
		if name == "" {
			result := dto.ImportRowResult{Row: rowNum, Status: "skipped", Error: "empty name — row skipped"}
			summary.Rows = append(summary.Rows, result)
			summary.Skipped++
			continue
		}

		req := dto.CreateCategoryRequest{
			Name:        name,
			Slug:        strings.TrimSpace(cells[1]),
			Description: strings.TrimSpace(cells[2]),
			IsActive:    true,
		}

		if v := strings.TrimSpace(cells[3]); v != "" {
			if id, err := strconv.ParseUint(v, 10, 64); err == nil {
				uid := uint(id)
				req.ParentID = &uid
			}
		}
		if v := strings.ToLower(strings.TrimSpace(cells[4])); v == "false" || v == "0" || v == "no" {
			req.IsActive = false
		}

		if err := s.validate.Struct(req); err != nil {
			result := dto.ImportRowResult{Row: rowNum, Status: "failed", Name: name, Error: err.Error()}
			summary.Rows = append(summary.Rows, result)
			summary.Failed++
			continue
		}

		created, err := s.categorySvc.Create(req)
		if err != nil {
			result := dto.ImportRowResult{Row: rowNum, Status: "failed", Name: name, Error: serviceErrMsg(err)}
			summary.Rows = append(summary.Rows, result)
			summary.Failed++
			continue
		}

		result := dto.ImportRowResult{Row: rowNum, Status: "created", Name: name, ID: &created.ID}
		summary.Rows = append(summary.Rows, result)
		summary.Created++
	}
	return summary, nil
}

// ─── Templates ───────────────────────────────────────────────────────────────

func (s *Service) ProductTemplate() ([]byte, error) {
	return buildTemplate("Products", productColumns, [][]string{
		{"Winter Jacket", "WJ-001", "99.99", "50", "Warm jacket for winter", "active", "", "", "", "129.99", "WJ001BAR", "5", "0.8"},
	})
}

func (s *Service) CategoryTemplate() ([]byte, error) {
	return buildTemplate("Categories", categoryColumns, [][]string{
		{"Clothing", "clothing", "All clothing items", "", "true"},
		{"Men's Jackets", "mens-jackets", "Jackets for men", "", "true"},
	})
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// padCells ensures the row slice has at least n elements.
func padCells(row []string, n int) []string {
	for len(row) < n {
		row = append(row, "")
	}
	return row
}

// serviceErrMsg extracts a human-readable message from a service error.
func serviceErrMsg(err error) string {
	if err == nil {
		return ""
	}
	if appErr, ok := err.(*utils.AppError); ok {
		return appErr.Message
	}
	return err.Error()
}

// buildTemplate creates an Excel workbook with a header row and optional example rows.
func buildTemplate(sheetName string, headers []string, examples [][]string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	_ = f.SetSheetName(sheet, sheetName)

	// Header style: bold + light blue background
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"1F6FEB"}, Pattern: 1},
	})

	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		_ = f.SetCellValue(sheetName, cell, h)
		_ = f.SetCellStyle(sheetName, cell, cell, style)
		_ = f.SetColWidth(sheetName, colLetter(col+1), colLetter(col+1), 18)
	}

	for rowIdx, row := range examples {
		for col, val := range row {
			cell, _ := excelize.CoordinatesToCellName(col+1, rowIdx+2)
			_ = f.SetCellValue(sheetName, cell, val)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// colLetter converts a 1-based column index to a letter (1→"A", 26→"Z", 27→"AA").
func colLetter(n int) string {
	name, _ := excelize.ColumnNumberToName(n)
	return name
}
