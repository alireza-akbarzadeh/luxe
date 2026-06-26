// package websocket holds the constants for the websocket package.
package websocket

import "fmt"

const (
	SalesFeedRoom = "admin_sales_feed"

	EventVendorOrderNew      = "vendor_order_new"
	EventVendorOrderUpdate   = "vendor_order_update"
	EventVendorOrderShipment = "vendor_order_shipment"
	EventVendorMessage       = "vendor_message"
)

// StoreRoom returns the WebSocket room id for a vendor store dashboard.
func StoreRoom(storeID uint) string {
	return fmt.Sprintf("store_%d", storeID)
}

// VendorUserRoom returns the room for vendor-wide events (messages, alerts).
func VendorUserRoom(userID uint) string {
	return fmt.Sprintf("vendor_%d", userID)
}
