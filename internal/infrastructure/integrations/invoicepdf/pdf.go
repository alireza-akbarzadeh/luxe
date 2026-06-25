package invoicepdf

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/go-pdf/fpdf"
)

// Generate builds a simple invoice PDF from the admin detail view.
func Generate(detail dto.InvoiceDetailResponse) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(18, 18, 18)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(0, 10, "INVOICE", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 6, fmt.Sprintf("Invoice #: %s", detail.InvoiceNumber), "", 1, "L", false, 0, "")
	if detail.OrderNumber != "" {
		pdf.CellFormat(0, 6, fmt.Sprintf("Order #: %s", detail.OrderNumber), "", 1, "L", false, 0, "")
	}
	pdf.CellFormat(0, 6, fmt.Sprintf("Status: %s", strings.ToUpper(detail.Status)), "", 1, "L", false, 0, "")
	if detail.IssuedAt != nil {
		pdf.CellFormat(0, 6, fmt.Sprintf("Issued: %s", detail.IssuedAt.Format(time.RFC1123)), "", 1, "L", false, 0, "")
	}
	pdf.Ln(4)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 8, "Bill to", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 6, detail.BillingName, "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, detail.BillingEmail, "", 1, "L", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(78, 8, "Item", "1", 0, "L", true, 0, "")
	pdf.CellFormat(22, 8, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Unit", "1", 0, "R", true, 0, "")
	pdf.CellFormat(35, 8, "Total", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	for _, item := range detail.Items {
		name := item.Name
		if name == "" {
			name = "Product"
		}
		pdf.CellFormat(78, 7, truncate(name, 42), "1", 0, "L", false, 0, "")
		pdf.CellFormat(22, 7, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 7, fmt.Sprintf("%.2f %s", item.UnitPrice, detail.Currency), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, fmt.Sprintf("%.2f %s", item.TotalPrice, detail.Currency), "1", 1, "R", false, 0, "")
	}

	pdf.Ln(4)
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(135, 7, "Subtotal", "", 0, "R", false, 0, "")
	pdf.CellFormat(35, 7, fmt.Sprintf("%.2f %s", detail.Subtotal, detail.Currency), "", 1, "R", false, 0, "")
	pdf.CellFormat(135, 7, "Shipping", "", 0, "R", false, 0, "")
	pdf.CellFormat(35, 7, fmt.Sprintf("%.2f %s", detail.ShippingAmount, detail.Currency), "", 1, "R", false, 0, "")
	pdf.CellFormat(135, 7, "Tax", "", 0, "R", false, 0, "")
	pdf.CellFormat(35, 7, fmt.Sprintf("%.2f %s", detail.TaxAmount, detail.Currency), "", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(135, 8, "Total", "", 0, "R", false, 0, "")
	pdf.CellFormat(35, 8, fmt.Sprintf("%.2f %s", detail.TotalAmount, detail.Currency), "", 1, "R", false, 0, "")

	if detail.Notes != "" {
		pdf.Ln(4)
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(0, 7, "Notes", "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 6, detail.Notes, "", "L", false)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max-3] + "..."
}
