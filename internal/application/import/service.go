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

// productColumn describes one Excel header used for bulk product create.
type productColumn struct {
	Key         string
	Required    bool
	Description string
	Example     string
}

// productColumns is the exact header order required by ImportProductsFromExcel.
var productColumns = []productColumn{
	{Key: "name", Required: true, Description: "Product display name (min 3 chars)", Example: "Relaxed Linen Button-Up"},
	{Key: "sku", Required: true, Description: "Unique SKU (3–50 chars)", Example: "LUX-W-TOP-003"},
	{Key: "price", Required: true, Description: "Sell price (number ≥ 0)", Example: "185"},
	{Key: "stock", Required: false, Description: "On-hand quantity (integer ≥ 0)", Example: "40"},
	{Key: "description", Required: false, Description: "Short product description", Example: "Oversized linen button-up with curved hem"},
	{Key: "status", Required: false, Description: "draft | active | inactive | archived (default: active)", Example: "active"},
	{Key: "category_id", Required: false, Description: "Existing category ID (number)", Example: "12"},
	{Key: "brand_id", Required: false, Description: "Existing brand ID (number)", Example: "3"},
	{Key: "store_id", Required: false, Description: "Existing store ID (number); falls back to query store_id", Example: "1"},
	{Key: "compare_at_price", Required: false, Description: "Compare-at / MSRP price", Example: "210"},
	{Key: "barcode", Required: false, Description: "Barcode / UPC", Example: "8901234567890"},
	{Key: "low_stock_threshold", Required: false, Description: "Low-stock alert threshold", Example: "5"},
	{Key: "weight", Required: false, Description: "Weight in kg", Example: "0.35"},
}

func productColumnKeys() []string {
	keys := make([]string, len(productColumns))
	for i, col := range productColumns {
		keys[i] = col.Key
	}
	return keys
}

func (s *Service) ImportProductsFromExcel(r io.Reader, storeID uint) (*dto.ImportSummary, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, utils.ErrBadRequest("invalid Excel file: " + err.Error())
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	for _, name := range f.GetSheetList() {
		if strings.EqualFold(name, "Products") {
			sheet = name
			break
		}
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, utils.ErrBadRequest("cannot read sheet: " + err.Error())
	}
	if len(rows) < 2 {
		return nil, utils.ErrBadRequest("file has no data rows (expected header + at least one data row)")
	}

	if err := validateProductHeader(rows[0]); err != nil {
		return nil, err
	}

	summary := &dto.ImportSummary{
		TotalRows: len(rows) - 1,
		Rows:      make([]dto.ImportRowResult, 0, len(rows)-1),
	}

	for i, row := range rows[1:] {
		rowNum := i + 2
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

		stock, _ := strconv.Atoi(strings.TrimSpace(cells[3]))
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

func validateProductHeader(header []string) error {
	keys := productColumnKeys()
	if len(header) < len(keys) {
		return utils.ErrBadRequest(fmt.Sprintf(
			"header must include columns in order: %s",
			strings.Join(keys, ", "),
		))
	}
	for i, key := range keys {
		got := strings.ToLower(strings.TrimSpace(header[i]))
		if got != key {
			return utils.ErrBadRequest(fmt.Sprintf(
				"column %d must be %q (got %q). Download the template for the exact header row.",
				i+1, key, header[i],
			))
		}
	}
	return nil
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
	keys := productColumnKeys()
	examples := [][]string{
		{
			"Relaxed Linen Button-Up", "LUX-W-TOP-003", "185", "40",
			"Oversized linen button-up with curved hem", "active",
			"", "", "", "210", "8901234567890", "5", "0.35",
		},
		{
			"Merino Crew Neck Sweater", "LUX-W-KNT-014", "220", "25",
			"Soft merino crew neck in seasonal colors", "active",
			"", "", "", "260", "", "3", "0.45",
		},
		{
			"Draft Sample Product", "LUX-DRAFT-001", "49.99", "0",
			"Replace IDs and publish when ready", "draft",
			"", "", "", "", "", "2", "",
		},
	}
	return buildProductTemplate(keys, examples)
}

func (s *Service) CategoryTemplate() ([]byte, error) {
	return buildTemplate("Categories", categoryColumns, [][]string{
		{"Clothing", "clothing", "All clothing items", "", "true"},
		{"Men's Jackets", "mens-jackets", "Jackets for men", "", "true"},
	})
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func padCells(row []string, n int) []string {
	for len(row) < n {
		row = append(row, "")
	}
	return row
}

func serviceErrMsg(err error) string {
	if err == nil {
		return ""
	}
	if appErr, ok := err.(*utils.AppError); ok {
		// Prefer the public message (already mapped for FK/unique); fall back to cause for unknowns.
		if appErr.Message != "" && appErr.Message != constants.ErrInternalServer.Error() {
			return appErr.Message
		}
		if appErr.Err != nil {
			return appErr.Err.Error()
		}
		return appErr.Message
	}
	return err.Error()
}

func buildProductTemplate(headers []string, examples [][]string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	const productsSheet = "Products"
	const guideSheet = "Column guide"

	defaultSheet := f.GetSheetName(0)
	if err := f.SetSheetName(defaultSheet, productsSheet); err != nil {
		return nil, err
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"1F6FEB"}, Pattern: 1},
	})
	requiredStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"8B6914"}, Pattern: 1},
	})

	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		_ = f.SetCellValue(productsSheet, cell, h)
		style := headerStyle
		if productColumns[col].Required {
			style = requiredStyle
		}
		_ = f.SetCellStyle(productsSheet, cell, cell, style)
		_ = f.SetColWidth(productsSheet, colLetter(col+1), colLetter(col+1), 20)
	}

	for rowIdx, row := range examples {
		for col, val := range row {
			cell, _ := excelize.CoordinatesToCellName(col+1, rowIdx+2)
			_ = f.SetCellValue(productsSheet, cell, val)
		}
	}
	_ = f.SetRowHeight(productsSheet, 1, 22)
	_ = f.AutoFilter(productsSheet, "A1:"+colLetter(len(headers))+"1", nil)

	if _, err := f.NewSheet(guideSheet); err != nil {
		return nil, err
	}
	guideHeaders := []string{"column", "required", "description", "example"}
	for col, h := range guideHeaders {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		_ = f.SetCellValue(guideSheet, cell, h)
		_ = f.SetCellStyle(guideSheet, cell, cell, headerStyle)
	}
	_ = f.SetColWidth(guideSheet, "A", "A", 22)
	_ = f.SetColWidth(guideSheet, "B", "B", 12)
	_ = f.SetColWidth(guideSheet, "C", "C", 55)
	_ = f.SetColWidth(guideSheet, "D", "D", 36)

	for i, col := range productColumns {
		row := i + 2
		_ = f.SetCellValue(guideSheet, fmt.Sprintf("A%d", row), col.Key)
		reqLabel := "optional"
		if col.Required {
			reqLabel = "required"
		}
		_ = f.SetCellValue(guideSheet, fmt.Sprintf("B%d", row), reqLabel)
		_ = f.SetCellValue(guideSheet, fmt.Sprintf("C%d", row), col.Description)
		_ = f.SetCellValue(guideSheet, fmt.Sprintf("D%d", row), col.Example)
	}

	noteRow := len(productColumns) + 3
	_ = f.SetCellValue(guideSheet, fmt.Sprintf("A%d", noteRow), "Notes")
	_ = f.SetCellValue(guideSheet, fmt.Sprintf("A%d", noteRow+1),
		"1) Keep the Products sheet header row exactly as provided (column order matters).")
	_ = f.SetCellValue(guideSheet, fmt.Sprintf("A%d", noteRow+2),
		"2) Required: name, sku, price. Status defaults to active when blank.")
	_ = f.SetCellValue(guideSheet, fmt.Sprintf("A%d", noteRow+3),
		"3) category_id / brand_id / store_id must reference existing IDs in your catalog.")
	_ = f.SetCellValue(guideSheet, fmt.Sprintf("A%d", noteRow+4),
		"4) Delete sample rows before importing your real catalog.")

	f.SetActiveSheet(0)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildTemplate(sheetName string, headers []string, examples [][]string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	defaultSheet := f.GetSheetName(0)
	if err := f.SetSheetName(defaultSheet, sheetName); err != nil {
		return nil, err
	}

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

func colLetter(n int) string {
	name, _ := excelize.ColumnNumberToName(n)
	return name
}
