package dto

import "time"

type AuditLogListFilters struct {
	Limit    int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset   int    `form:"offset" validate:"omitempty,min=0"`
	Search   string `form:"search" validate:"omitempty,max=200"`
	Action   string `form:"action" validate:"omitempty,max=32"`
	Resource string `form:"resource" validate:"omitempty,max=128"`
	UserID   uint   `form:"user_id" validate:"omitempty,min=1"`
	DateFrom string `form:"date_from" validate:"omitempty"`
	DateTo   string `form:"date_to" validate:"omitempty"`
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

type AuditLogSummaryResponse struct {
	Total        int64 `json:"total"`
	Last24Hours  int64 `json:"last_24_hours"`
	Today        int64 `json:"today"`
	UniqueActors int64 `json:"unique_actors"`
}
