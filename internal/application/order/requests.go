package order

import "github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"

// AdminOrderFilters extends public order filters with admin-only fields.
type AdminOrderFilters struct {
	dto.OrderFilters
	UserID *uint  `json:"user_id,omitempty"`
	Search string `json:"search,omitempty"`
}
