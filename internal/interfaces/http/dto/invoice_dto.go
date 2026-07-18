package dto

import (
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// AdminInvoiceListFilters supports admin listing of invoices.
type AdminInvoiceListFilters struct {
	Status   string `form:"status"`
	UserID   *uint  `form:"user_id"`
	OrderID  *uint  `form:"order_id"`
	Search   string `form:"search"`
	FromDate string `form:"from_date"`
	ToDate   string `form:"to_date"`
	Limit    int    `form:"limit"`
	Offset   int    `form:"offset"`
}

// AdminInvoiceListItem is a row in the admin invoices table.
type AdminInvoiceListItem struct {
	ID            uint    `json:"id"`
	InvoiceNumber string  `json:"invoice_number"`
	OrderID       uint    `json:"order_id"`
	OrderNumber   string  `json:"order_number,omitempty"`
	UserID        uint    `json:"user_id"`
	CustomerName  string  `json:"customer_name,omitempty"`
	CustomerEmail string  `json:"customer_email,omitempty"`
	TotalAmount   float64 `json:"total_amount"`
	Currency      string  `json:"currency"`
	Status        string  `json:"status"`
	IssuedAt      *string `json:"issued_at,omitempty"`
	PaidAt        *string `json:"paid_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

// AdminInvoiceListData wraps paginated admin invoice rows.
type AdminInvoiceListData struct {
	Invoices []AdminInvoiceListItem `json:"invoices"`
	Total    int64                  `json:"total"`
	Limit    int                    `json:"limit"`
	Offset   int                    `json:"offset"`
}

// InvoiceDetailResponse powers the admin invoice detail page.
type InvoiceDetailResponse struct {
	ID             uint                 `json:"id"`
	InvoiceNumber  string               `json:"invoice_number"`
	OrderID        uint                 `json:"order_id"`
	OrderNumber    string               `json:"order_number,omitempty"`
	UserID         uint                 `json:"user_id"`
	Status         string               `json:"status"`
	Subtotal       float64              `json:"subtotal"`
	TaxAmount      float64              `json:"tax_amount"`
	ShippingAmount float64              `json:"shipping_amount"`
	TotalAmount    float64              `json:"total_amount"`
	Currency       string               `json:"currency"`
	BillingName    string               `json:"billing_name,omitempty"`
	BillingEmail   string               `json:"billing_email,omitempty"`
	Notes          string               `json:"notes,omitempty"`
	PaymentStatus  string               `json:"payment_status,omitempty"`
	PaymentMethod  string               `json:"payment_method,omitempty"`
	IssuedAt       *time.Time           `json:"issued_at,omitempty"`
	DueAt          *time.Time           `json:"due_at,omitempty"`
	PaidAt         *time.Time           `json:"paid_at,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	Items          []AdminOrderItemView `json:"items"`
}

// UpdateInvoiceStatusRequest updates an invoice status (admin).
type UpdateInvoiceStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft issued paid void refunded"`
}

// ToAdminInvoiceListItem maps a preloaded invoice to a list row.
func ToAdminInvoiceListItem(inv *models.Invoice) AdminInvoiceListItem {
	item := AdminInvoiceListItem{
		ID:            inv.ID,
		InvoiceNumber: inv.InvoiceNumber,
		OrderID:       inv.OrderID,
		UserID:        inv.UserID,
		TotalAmount:   inv.TotalAmount,
		Currency:      inv.Currency,
		Status:        inv.Status,
		CreatedAt:     inv.CreatedAt.Format(time.RFC3339),
	}
	if inv.Order.ID != 0 {
		item.OrderNumber = inv.Order.OrderNumber
	}
	if inv.User.ID != 0 {
		item.CustomerName = strings.TrimSpace(inv.User.FirstName + " " + inv.User.LastName)
		item.CustomerEmail = inv.User.Email
	}
	if inv.IssuedAt != nil {
		formatted := inv.IssuedAt.Format(time.RFC3339)
		item.IssuedAt = &formatted
	}
	if inv.PaidAt != nil {
		formatted := inv.PaidAt.Format(time.RFC3339)
		item.PaidAt = &formatted
	}
	return item
}

// ToInvoiceDetail maps a fully preloaded invoice to the detail response.
func ToInvoiceDetail(inv *models.Invoice) InvoiceDetailResponse {
	resp := InvoiceDetailResponse{
		ID:             inv.ID,
		InvoiceNumber:  inv.InvoiceNumber,
		OrderID:        inv.OrderID,
		UserID:         inv.UserID,
		Status:         inv.Status,
		Subtotal:       inv.Subtotal,
		TaxAmount:      inv.TaxAmount,
		ShippingAmount: inv.ShippingAmount,
		TotalAmount:    inv.TotalAmount,
		Currency:       inv.Currency,
		BillingName:    inv.BillingName,
		BillingEmail:   inv.BillingEmail,
		Notes:          inv.Notes,
		IssuedAt:       inv.IssuedAt,
		DueAt:          inv.DueAt,
		PaidAt:         inv.PaidAt,
		CreatedAt:      inv.CreatedAt,
		UpdatedAt:      inv.UpdatedAt,
		Items:          []AdminOrderItemView{},
	}
	if inv.Order.ID != 0 {
		resp.OrderNumber = inv.Order.OrderNumber
		resp.Items = make([]AdminOrderItemView, len(inv.Order.Items))
		for i, orderItem := range inv.Order.Items {
			image := ""
			name := "Product"
			sku := ""
			category := ""
			if orderItem.Product.ID != 0 {
				name = orderItem.Product.Name
				sku = orderItem.Product.SKU
				if len(orderItem.Product.Images) > 0 {
					image = orderItem.Product.Images[0]
				}
				if orderItem.Product.Category.Name != "" {
					category = orderItem.Product.Category.Name
				}
			}
			resp.Items[i] = AdminOrderItemView{
				ID:         orderItem.ID,
				ProductID:  orderItem.ProductID,
				Name:       name,
				SKU:        sku,
				Image:      image,
				Quantity:   orderItem.Quantity,
				UnitPrice:  orderItem.Price,
				TotalPrice: orderItem.Total,
				Category:   category,
			}
		}
	}
	if inv.Payment != nil {
		resp.PaymentStatus = inv.Payment.Status
		resp.PaymentMethod = inv.Payment.Method
	} else {
		resp.PaymentStatus = constants.PaymentStatusPending
	}
	if resp.BillingName == "" && inv.User.ID != 0 {
		resp.BillingName = strings.TrimSpace(inv.User.FirstName + " " + inv.User.LastName)
	}
	if resp.BillingEmail == "" && inv.User.ID != 0 {
		resp.BillingEmail = inv.User.Email
	}
	return resp
}
