package notification

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/websocket"
)

// VendorStoreNotifier pushes store-scoped realtime events to vendor dashboards.
type VendorStoreNotifier interface {
	NotifyVendorStoresForOrder(
		ctx context.Context,
		orderID uint,
		wsEventType, notifType, title, message string,
		data map[string]interface{},
	)
}

// NotifyVendorStoresForOrder broadcasts to each affected store room and notifies store owners.
func (s *Service) NotifyVendorStoresForOrder(
	ctx context.Context,
	orderID uint,
	wsEventType, notifType, title, message string,
	data map[string]interface{},
) {
	if s.wsHub == nil || s.orderRepo == nil {
		return
	}

	storeIDs, err := s.orderRepo.ListStoreIDsForOrder(ctx, orderID)
	if err != nil || len(storeIDs) == 0 {
		return
	}

	if data == nil {
		data = map[string]interface{}{}
	}
	data["order_id"] = orderID

	for _, storeID := range storeIDs {
		payload := map[string]interface{}{}
		for k, v := range data {
			payload[k] = v
		}
		payload["store_id"] = storeID

		roomID := websocket.StoreRoom(storeID)
		s.BroadcastToRoom(roomID, wsEventType, payload)

		if s.storeRepo == nil {
			continue
		}
		store, err := s.storeRepo.FindByID(storeID)
		if err != nil || store.UserID == nil || *store.UserID == 0 {
			continue
		}
		ownerID := *store.UserID
		go func(uid uint) {
			_ = s.CreateNotification(uid, notifType, title, message, payload)
		}(ownerID)
	}
}
