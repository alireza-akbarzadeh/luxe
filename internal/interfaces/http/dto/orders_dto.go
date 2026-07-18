package dto

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// ConfirmCheckoutStripeRequest confirms order payment after Stripe Checkout redirect.
type ConfirmCheckoutStripeRequest struct {
	SessionID string `json:"session_id" validate:"required"`
}

// CheckoutResult is returned from checkout; MarshalJSON flattens the order for backward-compatible API responses.
type CheckoutResult struct {
	Order           *models.Order
	CheckoutURL     string
	StripeSessionID string
}

func (r CheckoutResult) MarshalJSON() ([]byte, error) {
	if r.Order == nil {
		return json.Marshal(map[string]interface{}{})
	}
	raw, err := json.Marshal(r.Order)
	if err != nil {
		return nil, err
	}
	payload := map[string]interface{}{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	if r.CheckoutURL != "" {
		payload["checkout_url"] = r.CheckoutURL
		payload["stripe_session_id"] = r.StripeSessionID
	}
	return json.Marshal(payload)
}

type CheckoutRequest struct {
	CouponCode     string `json:"coupon_code,omitempty"`
	Email          string `json:"email" validate:"required,email"`
	FirstName      string `json:"first_name" validate:"required"`
	LastName       string `json:"last_name" validate:"required"`
	AddressLine1   string `json:"address_line1" validate:"required"`
	AddressLine2   string `json:"address_line2"`
	City           string `json:"city" validate:"required"`
	State          string `json:"state" validate:"required"`
	Zip            string `json:"zip" validate:"required"`
	Country        string `json:"country" validate:"required"`
	Phone          string `json:"phone" validate:"required"`
	ShippingMethod string `json:"shipping_method"`
	PaymentMethod  string `json:"payment_method" validate:"omitempty,oneof=mock stripe wallet"`
	SaveInfo       bool   `json:"save_info"`
	Newsletter     bool   `json:"newsletter"`
	CardLast4      string `json:"card_last4,omitempty"`

	ShippingProviderID *uint  `json:"shipping_provider_id,omitempty"`
	CardNumber         string `json:"card_number" validate:"required_if=PaymentMethod mock,omitempty,len=16"`
	ExpiryMonth        int    `json:"expiry_month" validate:"required_if=PaymentMethod mock,omitempty,min=1,max=12"`
	ExpiryYear         int    `json:"expiry_year" validate:"required_if=PaymentMethod mock,omitempty,min=2025"`
	CVV                string `json:"cvv" validate:"required_if=PaymentMethod mock,omitempty,len=3"`
}

// NormalizePaymentMethod sets a default payment method based on Stripe availability.
func (r *CheckoutRequest) NormalizePaymentMethod(stripeEnabled bool) {
	if r.PaymentMethod != "" {
		return
	}
	if stripeEnabled {
		r.PaymentMethod = "stripe"
	} else {
		r.PaymentMethod = "mock"
	}
}

// CardInfo used internally (no JSON tags needed)
type CardInfo struct {
	CardNumber  string
	ExpiryMonth int
	ExpiryYear  int
	CVV         string
}

func MapAddress(userID uint, req CheckoutRequest) models.Address {
	recipientName := req.FirstName + " " + req.LastName
	return models.Address{
		UserID:        userID,
		AddressType:   "shipping",
		IsDefault:     req.SaveInfo,
		RecipientName: recipientName,
		Phone:         req.Phone,
		AddressLine1:  req.AddressLine1,
		AddressLine2:  req.AddressLine2,
		City:          req.City,
		State:         req.State,
		PostalCode:    req.Zip,
		Country:       req.Country,
		Instructions:  "",
	}
}

type OrderResponse struct {
	ID          uint      `json:"id"`
	OrderNumber string    `json:"order_number"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// AdminOrderListItem is a row in the admin orders table.
type AdminOrderListItem struct {
	ID             uint       `json:"id"`
	OrderNumber    string     `json:"order_number"`
	Status         string     `json:"status"`
	PaymentStatus  string     `json:"payment_status"`
	ShipmentStatus string     `json:"shipment_status,omitempty"`
	WorkflowState  *StateView `json:"workflow_state,omitempty"`
	Tags           []string   `json:"tags,omitempty"`
	TotalAmount    float64    `json:"total_amount"`
	Currency       string     `json:"currency"`
	CustomerName   string     `json:"customer_name"`
	CustomerEmail  string     `json:"customer_email"`
	ItemsCount     int        `json:"items_count"`
	CreatedAt      time.Time  `json:"created_at"`
}

// AdminOrderListData wraps paginated admin order rows.
type AdminOrderListData struct {
	Orders []AdminOrderListItem `json:"orders"`
	Total  int64                `json:"total"`
	Limit  int                  `json:"limit"`
	Offset int                  `json:"offset"`
}

// ToAdminOrderListItem maps an order model (with User preloaded) to an admin list row.
func ToAdminOrderListItem(order models.Order) AdminOrderListItem {
	name := strings.TrimSpace(order.User.FirstName + " " + order.User.LastName)
	if name == "" {
		name = order.User.Email
	}

	paymentStatus := constants.PaymentStatusPending
	if order.Payment != nil && order.Payment.Status != "" {
		paymentStatus = order.Payment.Status
	}

	shipmentStatus := ""
	if order.Shipment != nil {
		shipmentStatus = order.Shipment.Status
	}

	item := AdminOrderListItem{
		ID:             order.ID,
		OrderNumber:    order.OrderNumber,
		Status:         order.Status,
		PaymentStatus:  paymentStatus,
		ShipmentStatus: shipmentStatus,
		Tags:           orderTagStrings(order.Tags),
		TotalAmount:    order.TotalAmount,
		Currency:       order.Currency,
		CustomerName:   name,
		CustomerEmail:  order.User.Email,
		ItemsCount:     len(order.Items),
		CreatedAt:      order.CreatedAt,
	}
	if order.WorkflowState != nil {
		item.WorkflowState = ToStateView(order.WorkflowState)
	}
	return item
}

func orderTagStrings(tags []models.OrderTag) []string {
	if len(tags) == 0 {
		return nil
	}
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag.Tag != "" {
			out = append(out, tag.Tag)
		}
	}
	return out
}

// ToAdminOrderListItems maps orders to admin list rows.
func ToAdminOrderListItems(orders []models.Order) []AdminOrderListItem {
	items := make([]AdminOrderListItem, len(orders))
	for i, order := range orders {
		items[i] = ToAdminOrderListItem(order)
	}
	return items
}

// AdminOrderItemView is a line item on the admin order detail page.
type AdminOrderItemView struct {
	ID         uint    `json:"id"`
	ProductID  uint    `json:"product_id"`
	Name       string  `json:"name"`
	SKU        string  `json:"sku"`
	Image      string  `json:"image,omitempty"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
	Category   string  `json:"category,omitempty"`
}

// AdminOrderShippingAddress is the shipment destination on an order.
type AdminOrderShippingAddress struct {
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2,omitempty"`
	City         string `json:"city"`
	State        string `json:"state,omitempty"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
}

// AdminOrderDetailResponse powers the admin order detail page.
type AdminOrderDetailResponse struct {
	ID                uint                       `json:"id"`
	OrderNumber       string                     `json:"order_number"`
	Status            string                     `json:"status"`
	PaymentStatus     string                     `json:"payment_status"`
	PaymentMethod     string                     `json:"payment_method,omitempty"`
	ShipmentStatus    string                     `json:"shipment_status,omitempty"`
	WorkflowState     *StateView                 `json:"workflow_state,omitempty"`
	Tags              []string                   `json:"tags,omitempty"`
	ParentOrderID     *uint                      `json:"parent_order_id,omitempty"`
	TotalAmount       float64                    `json:"total_amount"`
	Currency          string                     `json:"currency"`
	Notes             string                     `json:"notes,omitempty"`
	CustomerName      string                     `json:"customer_name"`
	CustomerEmail     string                     `json:"customer_email"`
	TrackingNumber    string                     `json:"tracking_number,omitempty"`
	Carrier           string                     `json:"carrier,omitempty"`
	EstimatedDelivery *time.Time                 `json:"estimated_delivery,omitempty"`
	ShippingAddress   *AdminOrderShippingAddress `json:"shipping_address,omitempty"`
	CreatedAt         time.Time                  `json:"created_at"`
	UpdatedAt         time.Time                  `json:"updated_at"`
	Items             []AdminOrderItemView       `json:"items"`
	Tracking          *OrderTrackingDetailView   `json:"tracking,omitempty"`
}

// UpdateOrderNotesRequest updates admin/customer notes on an order.
type UpdateOrderNotesRequest struct {
	Notes string `json:"notes" validate:"max=2000"`
}

// UpdateOrderTagsRequest replaces admin tags on an order.
type UpdateOrderTagsRequest struct {
	Tags []string `json:"tags" validate:"max=20,dive,max=64"`
}

// ToAdminOrderDetail maps a fully preloaded order to the admin detail response.
func ToAdminOrderDetail(order models.Order) AdminOrderDetailResponse {
	name := strings.TrimSpace(order.User.FirstName + " " + order.User.LastName)
	if name == "" {
		name = order.User.Email
	}

	paymentStatus := constants.PaymentStatusPending
	paymentMethod := ""
	if order.Payment != nil {
		if order.Payment.Status != "" {
			paymentStatus = order.Payment.Status
		}
		paymentMethod = order.Payment.Method
	}

	items := make([]AdminOrderItemView, len(order.Items))
	for i, item := range order.Items {
		image := ""
		name := "Product"
		sku := ""
		category := ""
		if item.Product.ID != 0 {
			name = item.Product.Name
			sku = item.Product.SKU
			if len(item.Product.Images) > 0 {
				image = item.Product.Images[0]
			}
			if item.Product.Category.Name != "" {
				category = item.Product.Category.Name
			}
		}
		items[i] = AdminOrderItemView{
			ID:         item.ID,
			ProductID:  item.ProductID,
			Name:       name,
			SKU:        sku,
			Image:      image,
			Quantity:   item.Quantity,
			UnitPrice:  item.Price,
			TotalPrice: item.Total,
			Category:   category,
		}
	}

	detail := AdminOrderDetailResponse{
		ID:            order.ID,
		OrderNumber:   order.OrderNumber,
		Status:        order.Status,
		PaymentStatus: paymentStatus,
		PaymentMethod: paymentMethod,
		Tags:          orderTagStrings(order.Tags),
		ParentOrderID: order.ParentOrderID,
		TotalAmount:   order.TotalAmount,
		Currency:      order.Currency,
		Notes:         order.Notes,
		CustomerName:  name,
		CustomerEmail: order.User.Email,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
		Items:         items,
	}
	if order.WorkflowState != nil {
		detail.WorkflowState = ToStateView(order.WorkflowState)
	}

	if order.Shipment != nil {
		detail.ShipmentStatus = order.Shipment.Status
		detail.TrackingNumber = order.Shipment.TrackingNumber
		detail.Carrier = order.Shipment.Carrier
		if order.Shipment.EstimatedDelivery != nil {
			detail.EstimatedDelivery = order.Shipment.EstimatedDelivery
		}
		detail.ShippingAddress = &AdminOrderShippingAddress{
			AddressLine1: order.Shipment.AddressLine1,
			AddressLine2: order.Shipment.AddressLine2,
			City:         order.Shipment.City,
			State:        order.Shipment.State,
			PostalCode:   order.Shipment.PostalCode,
			Country:      order.Shipment.Country,
		}
	}

	tracking := BuildOrderTrackingDetail(order, nil)
	detail.Tracking = &tracking

	return detail
}

// PerformOrderTransitionRequest triggers a workflow event on an order (admin or vendor).
type PerformOrderTransitionRequest struct {
	Event          string `json:"event" validate:"required,min=1,max=64"`
	Note           string `json:"note" validate:"omitempty,max=512"`
	TrackingNumber string `json:"tracking_number" validate:"omitempty,max=128"`
}

// OrderTransitionResponse is returned after a successful order workflow transition.
type OrderTransitionResponse struct {
	Transition TransitionResultView `json:"transition"`
	Order      interface{}          `json:"order"` // models.Order in responses
}

// VendorOrderListItem is a row in the vendor orders table (store-scoped).
type VendorOrderListItem struct {
	ID              uint      `json:"id"`
	OrderNumber     string    `json:"order_number"`
	Status          string    `json:"status"`
	PaymentStatus   string    `json:"payment_status"`
	PaymentMethod   string    `json:"payment_method,omitempty"`
	TotalAmount     float64   `json:"total_amount"`
	StoreSubtotal   float64   `json:"store_subtotal"`
	Currency        string    `json:"currency"`
	CustomerName    string    `json:"customer_name"`
	CustomerEmail   string    `json:"customer_email"`
	ItemsCount      int       `json:"items_count"`
	StoreItemsCount int       `json:"store_items_count"`
	TrackingNumber  string    `json:"tracking_number,omitempty"`
	Carrier         string    `json:"carrier,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// VendorOrderListData wraps paginated vendor order rows.
type VendorOrderListData struct {
	Orders []VendorOrderListItem `json:"orders"`
	Total  int64                 `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

// VendorOrderStatsResponse summarizes order counts for a vendor store dashboard.
type VendorOrderStatsResponse struct {
	Total    int64            `json:"total"`
	ByStatus map[string]int64 `json:"by_status"`
}

// VendorOrderDetailResponse powers the vendor order detail view (store line items only).
type VendorOrderDetailResponse struct {
	AdminOrderDetailResponse
	StoreSubtotal   float64 `json:"store_subtotal"`
	StoreItemsCount int     `json:"store_items_count"`
}

func vendorCustomerName(order models.Order) string {
	name := strings.TrimSpace(order.User.FirstName + " " + order.User.LastName)
	if name == "" {
		return order.User.Email
	}
	return name
}

func vendorPaymentMeta(order models.Order) (status, method string) {
	status = constants.PaymentStatusPending
	if order.Payment != nil {
		if order.Payment.Status != "" {
			status = order.Payment.Status
		}
		method = order.Payment.Method
	}
	return status, method
}

// ToVendorOrderListItem maps an order to a vendor list row for the given store.
func ToVendorOrderListItem(order models.Order, storeID uint) VendorOrderListItem {
	paymentStatus, paymentMethod := vendorPaymentMeta(order)

	storeItemsCount := 0
	storeSubtotal := 0.0
	for _, item := range order.Items {
		if item.Product.StoreID == storeID {
			storeItemsCount += item.Quantity
			storeSubtotal += item.Total
		}
	}

	item := VendorOrderListItem{
		ID:              order.ID,
		OrderNumber:     order.OrderNumber,
		Status:          order.Status,
		PaymentStatus:   paymentStatus,
		PaymentMethod:   paymentMethod,
		TotalAmount:     order.TotalAmount,
		StoreSubtotal:   storeSubtotal,
		Currency:        order.Currency,
		CustomerName:    vendorCustomerName(order),
		CustomerEmail:   order.User.Email,
		ItemsCount:      len(order.Items),
		StoreItemsCount: storeItemsCount,
		CreatedAt:       order.CreatedAt,
	}

	if order.Shipment != nil {
		item.TrackingNumber = order.Shipment.TrackingNumber
		item.Carrier = order.Shipment.Carrier
	}

	return item
}

// ToVendorOrderListItems maps orders to vendor list rows.
func ToVendorOrderListItems(orders []models.Order, storeID uint) []VendorOrderListItem {
	items := make([]VendorOrderListItem, len(orders))
	for i, order := range orders {
		items[i] = ToVendorOrderListItem(order, storeID)
	}
	return items
}

// ToVendorOrderDetail maps a fully preloaded order to vendor detail (store items only).
func ToVendorOrderDetail(order models.Order, storeID uint) VendorOrderDetailResponse {
	base := ToAdminOrderDetail(order)

	filtered := make([]AdminOrderItemView, 0, len(base.Items))
	storeSubtotal := 0.0
	storeItemsCount := 0
	for _, item := range order.Items {
		if item.Product.StoreID != storeID {
			continue
		}
		storeItemsCount += item.Quantity
		storeSubtotal += item.Total

		view := AdminOrderItemView{
			ID:         item.ID,
			ProductID:  item.ProductID,
			Name:       item.Product.Name,
			SKU:        item.Product.SKU,
			Quantity:   item.Quantity,
			UnitPrice:  item.Price,
			TotalPrice: item.Total,
		}
		if len(item.Product.Images) > 0 {
			view.Image = item.Product.Images[0]
		}
		if item.Product.Category.Name != "" {
			view.Category = item.Product.Category.Name
		}
		filtered = append(filtered, view)
	}

	base.Items = filtered
	return VendorOrderDetailResponse{
		AdminOrderDetailResponse: base,
		StoreSubtotal:            storeSubtotal,
		StoreItemsCount:          storeItemsCount,
	}
}
