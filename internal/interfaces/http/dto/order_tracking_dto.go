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

type orderTrackingMilestoneDef struct {
	key         string
	title       string
	description string
}

var orderTrackingMilestoneDefs = []orderTrackingMilestoneDef{
	{key: "order_confirmed", title: "Order Confirmed", description: "We received your order"},
	{key: "payment_received", title: "Payment Received", description: "Payment verified successfully"},
	{key: "warehouse_processing", title: "Warehouse Processing", description: "Items picked from inventory"},
	{key: "quality_inspection", title: "Quality Inspection", description: "Products checked before packing"},
	{key: "packaged", title: "Packaged", description: "Order sealed and labeled"},
	{key: "shipped", title: "Shipped", description: "Handed to carrier"},
	{key: "out_for_delivery", title: "Out for Delivery", description: "Driver is on the way"},
	{key: "delivered", title: "Delivered", description: "Package received"},
}

// BuildOrderTrackingDetail synthesizes tracking UI data from a preloaded order.
func BuildOrderTrackingDetail(order models.Order) OrderTrackingDetailView {
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

	activeIndex := orderTrackingActiveIndex(status, paymentStatus, shipmentStatus)
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

	milestones := buildOrderTrackingMilestones(activeIndex, order.CreatedAt, paymentAt, shippedAt, deliveredAt, estimatedDelivery)
	events := buildOrderTrackingEvents(order, status, carrier, trackingNumber, paymentAt, shippedAt, deliveredAt)

	statusLabel := orderTrackingStatusLabel(status, shipmentStatus)
	progressPercent := orderTrackingProgressPercent(activeIndex)

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
	if activeIndex >= 6 && status != constants.OrderStatusDelivered {
		driver = &OrderTrackingDriverView{
			Name:             "Michael Brown",
			Rating:           4.9,
			Carrier:          carrier,
			Vehicle:          carrier + " Delivery Van",
			LicensePlate:     "DHL-7842",
			EstimatedArrival: formatEstimatedArrival(estimatedDelivery),
		}
	}

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

func orderTrackingActiveIndex(orderStatus, paymentStatus, shipmentStatus string) int {
	switch orderStatus {
	case constants.OrderStatusDelivered:
		return 8
	case constants.OrderStatusCancelled, constants.OrderStatusRefunded:
		return 0
	case constants.OrderStatusShipped:
		if shipmentStatus == "out_for_delivery" || shipmentStatus == "in_transit" {
			return 7
		}
		return 6
	case constants.OrderStatusPaid, constants.OrderStatusDelayed:
		if paymentStatus == constants.PaymentStatusSucceeded || paymentStatus == constants.PaymentStatusCompleted {
			return 3
		}
		return 2
	case constants.OrderStatusPending:
		if paymentStatus == constants.PaymentStatusSucceeded || paymentStatus == constants.PaymentStatusCompleted {
			return 2
		}
		return 1
	default:
		return 1
	}
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

func orderTrackingProgressPercent(activeIndex int) int {
	if activeIndex <= 0 {
		return 8
	}
	if activeIndex >= len(orderTrackingMilestoneDefs) {
		return 100
	}
	return int(float64(activeIndex) / float64(len(orderTrackingMilestoneDefs)) * 100)
}

func buildOrderTrackingMilestones(
	activeIndex int,
	createdAt, paymentAt time.Time,
	shippedAt, deliveredAt, estimatedDelivery *time.Time,
) []OrderTrackingMilestoneView {
	milestones := make([]OrderTrackingMilestoneView, len(orderTrackingMilestoneDefs))
	for i, def := range orderTrackingMilestoneDefs {
		step := i + 1
		status := "upcoming"
		if step < activeIndex {
			status = "completed"
		} else if step == activeIndex {
			status = "active"
		}

		var occurredAt *time.Time
		switch def.key {
		case "order_confirmed":
			t := createdAt
			occurredAt = &t
		case "payment_received":
			t := paymentAt
			occurredAt = &t
		case "warehouse_processing", "quality_inspection", "packaged":
			if activeIndex > step {
				t := paymentAt.Add(time.Duration(step) * time.Hour)
				occurredAt = &t
			}
		case "shipped":
			occurredAt = shippedAt
		case "out_for_delivery":
			if shippedAt != nil {
				t := shippedAt.Add(6 * time.Hour)
				occurredAt = &t
			}
		case "delivered":
			occurredAt = deliveredAt
			if occurredAt == nil && estimatedDelivery != nil && activeIndex >= 8 {
				occurredAt = estimatedDelivery
			}
		}

		milestones[i] = OrderTrackingMilestoneView{
			Key:         def.key,
			Title:       def.title,
			Description: def.description,
			Status:      status,
			OccurredAt:  occurredAt,
		}
	}
	return milestones
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
