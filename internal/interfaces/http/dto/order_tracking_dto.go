package dto

import (
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// OrderTrackingMilestoneView is one step in the delivery progress timeline.
type OrderTrackingMilestoneView struct {
	Key         string     `json:"key"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      string     `json:"status"`
	OccurredAt  *time.Time `json:"occurred_at,omitempty"`
}

// OrderTrackingEventView is an activity log entry for the order.
type OrderTrackingEventView struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// OrderTrackingDeliveryView is shipment destination and package metadata.
type OrderTrackingDeliveryView struct {
	RecipientName     string   `json:"recipient_name"`
	Phone             string   `json:"phone,omitempty"`
	AddressLine1      string   `json:"address_line1,omitempty"`
	AddressLine2      string   `json:"address_line2,omitempty"`
	City              string   `json:"city,omitempty"`
	State             string   `json:"state,omitempty"`
	PostalCode        string   `json:"postal_code,omitempty"`
	Country           string   `json:"country,omitempty"`
	Instructions      string   `json:"instructions,omitempty"`
	ServiceName       string   `json:"service_name,omitempty"`
	PackageWeightKg   float64  `json:"package_weight_kg,omitempty"`
	PackageDimensions string   `json:"package_dimensions,omitempty"`
	InsuranceIncluded bool     `json:"insurance_included"`
	SignatureRequired bool     `json:"signature_required"`
	DestinationLat    *float64 `json:"destination_lat,omitempty"`
	DestinationLng    *float64 `json:"destination_lng,omitempty"`
	HubLat            *float64 `json:"hub_lat,omitempty"`
	HubLng            *float64 `json:"hub_lng,omitempty"`
	DistanceMiles     float64  `json:"distance_miles,omitempty"`
	StopsRemaining    int      `json:"stops_remaining,omitempty"`
}

// OrderTrackingPaymentSummaryView breaks down totals for the tracking page.
type OrderTrackingPaymentSummaryView struct {
	Subtotal      float64 `json:"subtotal"`
	Discount      float64 `json:"discount"`
	Shipping      float64 `json:"shipping"`
	Tax           float64 `json:"tax"`
	Total         float64 `json:"total"`
	Currency      string  `json:"currency"`
	Method        string  `json:"method,omitempty"`
	TransactionID string  `json:"transaction_id,omitempty"`
	CardLast4     string  `json:"card_last4,omitempty"`
}

// OrderTrackingDriverView is carrier-assigned delivery driver info (when in transit).
type OrderTrackingDriverView struct {
	Name             string  `json:"name,omitempty"`
	Rating           float64 `json:"rating,omitempty"`
	Carrier          string  `json:"carrier,omitempty"`
	Vehicle          string  `json:"vehicle,omitempty"`
	LicensePlate     string  `json:"license_plate,omitempty"`
	EstimatedArrival string  `json:"estimated_arrival,omitempty"`
}

// OrderTrackingCourierView summarizes the active carrier shipment.
type OrderTrackingCourierView struct {
	Name           string `json:"name"`
	TrackingNumber string `json:"tracking_number,omitempty"`
	Service        string `json:"service,omitempty"`
	TotalItems     int    `json:"total_items"`
}

// OrderTrackingDetailView powers the storefront order tracking page.
type OrderTrackingDetailView struct {
	StatusLabel      string                           `json:"status_label"`
	ProgressPercent  int                              `json:"progress_percent"`
	EstimatedArrival string                           `json:"estimated_arrival,omitempty"`
	Milestones       []OrderTrackingMilestoneView     `json:"milestones"`
	Events           []OrderTrackingEventView         `json:"events"`
	Delivery         *OrderTrackingDeliveryView       `json:"delivery,omitempty"`
	PaymentSummary   *OrderTrackingPaymentSummaryView `json:"payment_summary,omitempty"`
	Driver           *OrderTrackingDriverView         `json:"driver,omitempty"`
	Courier          *OrderTrackingCourierView        `json:"courier,omitempty"`
}

// BuildOrderTrackingDetail synthesizes tracking UI data from a preloaded order.
// Pass workflow states (sorted by SortOrder) to drive milestones — never hardcode steps.
func BuildOrderTrackingDetail(order models.Order, states []models.WorkflowState) OrderTrackingDetailView {
	status := strings.ToLower(strings.TrimSpace(order.Status))
	paymentStatus := constants.PaymentStatusPending
	paymentMethod := ""
	transactionID := ""
	paymentAt := order.CreatedAt

	if order.Payment != nil {
		if order.Payment.Status != "" {
			paymentStatus = order.Payment.Status
		}
		paymentMethod = order.Payment.Method
		transactionID = order.Payment.TransactionID
		paymentAt = order.Payment.CreatedAt
	}

	shipmentStatus := ""
	carrier := "Standard Shipping"
	trackingNumber := ""
	shippingPrice := 0.0
	var estimatedDelivery *time.Time
	var shippedAt *time.Time
	var deliveredAt *time.Time
	serviceName := "Standard Delivery"

	var shipment *models.Shipment
	if order.Shipment != nil {
		shipment = order.Shipment
		shipmentStatus = strings.ToLower(shipment.Status)
		if shipment.Carrier != "" {
			carrier = shipment.Carrier
		}
		trackingNumber = shipment.TrackingNumber
		shippingPrice = shipment.ShippingPrice
		estimatedDelivery = shipment.EstimatedDelivery
		shippedAt = shipment.ShippedAt
		deliveredAt = shipment.DeliveredAt
		if shipment.Provider != nil && shipment.Provider.Name != "" {
			serviceName = shipment.Provider.Name
		}
	}

	currentCode := orderTrackingCurrentCode(order, status)
	milestones := buildMilestonesFromWorkflowStates(states, currentCode)
	progressPercent := orderTrackingProgressFromMilestones(milestones)

	itemCount := 0
	subtotal := 0.0
	for _, item := range order.Items {
		itemCount += item.Quantity
		subtotal += item.Total
	}

	discount := 0.0
	if subtotal+shippingPrice > order.TotalAmount {
		discount = subtotal + shippingPrice - order.TotalAmount
	}
	tax := order.TotalAmount - subtotal - shippingPrice + discount
	if tax < 0 {
		tax = 0
	}

	events := buildOrderTrackingEvents(order, status, carrier, trackingNumber, paymentAt, shippedAt, deliveredAt)
	statusLabel := orderTrackingStatusLabel(status, shipmentStatus)
	if order.WorkflowState != nil && order.WorkflowState.Name != "" {
		statusLabel = order.WorkflowState.Name
	}

	recipientName := strings.TrimSpace(order.User.FirstName + " " + order.User.LastName)
	if recipientName == "" {
		recipientName = order.User.Email
	}

	delivery := &OrderTrackingDeliveryView{
		RecipientName:     recipientName,
		Phone:             order.User.Phone,
		Instructions:      strings.TrimSpace(order.Notes),
		ServiceName:       serviceName,
		PackageWeightKg:   estimatePackageWeightKg(itemCount),
		PackageDimensions: estimatePackageDimensions(itemCount),
		InsuranceIncluded: order.TotalAmount >= 100,
		SignatureRequired: order.TotalAmount >= 250,
		DistanceMiles:     2.4,
		StopsRemaining:    3,
	}

	if shipment != nil {
		delivery.AddressLine1 = shipment.AddressLine1
		delivery.AddressLine2 = shipment.AddressLine2
		delivery.City = shipment.City
		delivery.State = shipment.State
		delivery.PostalCode = shipment.PostalCode
		delivery.Country = shipment.Country
		delivery.DestinationLat, delivery.DestinationLng = geocodeHintCoords(shipment.City, shipment.State, shipment.Country)
		if delivery.DestinationLat != nil && delivery.DestinationLng != nil {
			hubLat := *delivery.DestinationLat + 0.04
			hubLng := *delivery.DestinationLng - 0.06
			delivery.HubLat = &hubLat
			delivery.HubLng = &hubLng
		}
	}

	courier := &OrderTrackingCourierView{
		Name:           carrier,
		TrackingNumber: trackingNumber,
		Service:        serviceName,
		TotalItems:     itemCount,
	}

	var driver *OrderTrackingDriverView
	if (currentCode == "shipped" || currentCode == "out_for_delivery" || currentCode == "in_transit" || status == constants.OrderStatusShipped) &&
		status != constants.OrderStatusDelivered {
		driver = &OrderTrackingDriverView{
			Name:             "Michael Brown",
			Rating:           4.9,
			Carrier:          carrier,
			Vehicle:          carrier + " Delivery Van",
			LicensePlate:     "DHL-7842",
			EstimatedArrival: formatEstimatedArrival(estimatedDelivery),
		}
	}

	_ = paymentStatus

	return OrderTrackingDetailView{
		StatusLabel:      statusLabel,
		ProgressPercent:  progressPercent,
		EstimatedArrival: formatEstimatedArrival(estimatedDelivery),
		Milestones:       milestones,
		Events:           events,
		Delivery:         delivery,
		PaymentSummary: &OrderTrackingPaymentSummaryView{
			Subtotal:      subtotal,
			Discount:      discount,
			Shipping:      shippingPrice,
			Tax:           tax,
			Total:         order.TotalAmount,
			Currency:      order.Currency,
			Method:        paymentMethod,
			TransactionID: transactionID,
			CardLast4:     cardLast4FromTransaction(transactionID),
		},
		Driver:  driver,
		Courier: courier,
	}
}

func orderTrackingCurrentCode(order models.Order, status string) string {
	if order.WorkflowState != nil && order.WorkflowState.Code != "" {
		return strings.ToLower(order.WorkflowState.Code)
	}
	switch status {
	case constants.OrderStatusPaid:
		return "paid"
	case constants.OrderStatusShipped:
		return "shipped"
	case constants.OrderStatusDelivered:
		return "delivered"
	case constants.OrderStatusCancelled:
		return "cancelled"
	case constants.OrderStatusRefunded:
		return "refunded"
	default:
		return "pending"
	}
}

func buildMilestonesFromWorkflowStates(states []models.WorkflowState, currentCode string) []OrderTrackingMilestoneView {
	filtered := make([]models.WorkflowState, 0, len(states))
	for _, s := range states {
		code := strings.ToLower(s.Code)
		if code == "cancelled" || code == "refunded" {
			continue
		}
		filtered = append(filtered, s)
	}

	// Sort by SortOrder ascending (stable copy)
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].SortOrder < filtered[i].SortOrder {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	activeIndex := 0
	for i, s := range filtered {
		if strings.ToLower(s.Code) == currentCode {
			activeIndex = i
			break
		}
	}
	isFinal := false
	if activeIndex < len(filtered) && filtered[activeIndex].IsFinal {
		isFinal = true
	}
	if currentCode == constants.OrderStatusDelivered {
		isFinal = true
	}

	milestones := make([]OrderTrackingMilestoneView, len(filtered))
	for i, s := range filtered {
		status := "upcoming"
		if isFinal || i < activeIndex {
			status = "completed"
		} else if i == activeIndex {
			status = "active"
		}
		milestones[i] = OrderTrackingMilestoneView{
			Key:         s.Code,
			Title:       s.Name,
			Description: s.Description,
			Status:      status,
		}
	}
	return milestones
}

func orderTrackingProgressFromMilestones(milestones []OrderTrackingMilestoneView) int {
	if len(milestones) == 0 {
		return 8
	}
	completed := 0
	active := 0
	for _, m := range milestones {
		if m.Status == "completed" {
			completed++
		}
		if m.Status == "active" {
			active++
		}
	}
	pct := int(float64(completed+active)*100.0/float64(len(milestones)) + 0.5)
	if pct < 8 {
		return 8
	}
	if pct > 100 {
		return 100
	}
	return pct
}

func orderTrackingStatusLabel(orderStatus, shipmentStatus string) string {
	switch orderStatus {
	case constants.OrderStatusDelivered:
		return "Delivered"
	case constants.OrderStatusShipped:
		if shipmentStatus == "out_for_delivery" || shipmentStatus == "in_transit" {
			return "In Transit"
		}
		return "Shipped"
	case constants.OrderStatusPaid:
		return "Processing"
	case constants.OrderStatusCancelled:
		return "Cancelled"
	case constants.OrderStatusRefunded:
		return "Refunded"
	default:
		return "Order Placed"
	}
}

func buildOrderTrackingEvents(
	order models.Order,
	status, carrier, trackingNumber string,
	paymentAt time.Time,
	shippedAt, deliveredAt *time.Time,
) []OrderTrackingEventView {
	events := []OrderTrackingEventView{
		{
			ID:        fmt.Sprintf("evt-confirmed-%d", order.ID),
			Type:      "order_confirmed",
			Title:     "Order confirmed",
			Message:   fmt.Sprintf("Order #%s was placed successfully.", order.OrderNumber),
			Timestamp: order.CreatedAt,
		},
	}

	if paymentAt.After(order.CreatedAt) {
		events = append([]OrderTrackingEventView{{
			ID:        fmt.Sprintf("evt-payment-%d", order.ID),
			Type:      "payment_received",
			Title:     "Payment received",
			Message:   "Your payment was processed successfully.",
			Timestamp: paymentAt,
		}}, events...)
	}

	if status == constants.OrderStatusPaid || status == constants.OrderStatusShipped || status == constants.OrderStatusDelivered {
		events = append([]OrderTrackingEventView{{
			ID:        fmt.Sprintf("evt-processing-%d", order.ID),
			Type:      "warehouse_processing",
			Title:     "Warehouse processing",
			Message:   "Your items are being prepared for shipment.",
			Timestamp: paymentAt.Add(2 * time.Hour),
		}}, events...)
	}

	if shippedAt != nil {
		msg := "Your package has left our facility."
		if trackingNumber != "" {
			msg = fmt.Sprintf("Tracking number %s is active with %s.", trackingNumber, carrier)
		}
		events = append([]OrderTrackingEventView{{
			ID:        fmt.Sprintf("evt-shipped-%d", order.ID),
			Type:      "shipped",
			Title:     "Out for delivery",
			Message:   msg,
			Timestamp: *shippedAt,
		}}, events...)
	}

	if deliveredAt != nil {
		events = append([]OrderTrackingEventView{{
			ID:        fmt.Sprintf("evt-delivered-%d", order.ID),
			Type:      "delivered",
			Title:     "Delivered",
			Message:   "Your order was delivered successfully.",
			Timestamp: *deliveredAt,
		}}, events...)
	}

	return events
}

func formatEstimatedArrival(estimated *time.Time) string {
	if estimated == nil {
		return ""
	}
	return estimated.Format("Monday, Jan 2 before 3:04 PM")
}

func estimatePackageWeightKg(itemCount int) float64 {
	if itemCount <= 0 {
		return 1.2
	}
	return float64(itemCount) * 1.1
}

func estimatePackageDimensions(itemCount int) string {
	if itemCount >= 3 {
		return "32 x 24 x 12 cm"
	}
	return "24 x 18 x 8 cm"
}

func cardLast4FromTransaction(transactionID string) string {
	trimmed := strings.TrimSpace(transactionID)
	if len(trimmed) < 4 {
		return ""
	}
	return trimmed[len(trimmed)-4:]
}

func geocodeHintCoords(city, state, country string) (*float64, *float64) {
	key := strings.ToLower(strings.TrimSpace(city + " " + state + " " + country))
	switch {
	case strings.Contains(key, "new york"):
		lat, lng := 40.7580, -73.9855
		return &lat, &lng
	case strings.Contains(key, "los angeles"):
		lat, lng := 34.0522, -118.2437
		return &lat, &lng
	case strings.Contains(key, "london"):
		lat, lng := 51.5074, -0.1278
		return &lat, &lng
	case strings.Contains(key, "tehran"):
		lat, lng := 35.6892, 51.3890
		return &lat, &lng
	default:
		lat, lng := 40.7128, -74.0060
		return &lat, &lng
	}
}
