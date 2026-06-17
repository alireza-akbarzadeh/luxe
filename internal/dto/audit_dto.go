package dto

import "time"

type AuditLogListFilters struct {
	Limit  int `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset int `form:"offset" validate:"omitempty,min=0"`
}

type AuditLogResponse struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id,omitempty"`
	Path       string    `json:"path"`
	IPAddress  string    `json:"ip_address,omitempty"`
	RequestID  string    `json:"request_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UserEmail  string    `json:"user_email,omitempty"`
}
