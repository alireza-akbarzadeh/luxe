package dto

import "time"

// CreateSupportTicketRequest opens a new support ticket (guest or authenticated).
type CreateSupportTicketRequest struct {
	Subject       string `json:"subject" validate:"required,min=1,max=255"`
	Message       string `json:"message" validate:"required,min=1,max=4096"`
	Channel       string `json:"channel" validate:"omitempty,oneof=email chat web"`
	OrderID       *uint  `json:"order_id,omitempty" validate:"omitempty,gt=0"`
	CustomerName  string `json:"customer_name" validate:"omitempty,max=200"`
	CustomerEmail string `json:"customer_email" validate:"omitempty,email,max=255"`
	Priority      string `json:"priority" validate:"omitempty,oneof=low normal high urgent"`
}

// AdminSupportTicketListFilters are query params for admin ticket listing.
type AdminSupportTicketListFilters struct {
	Search   string `form:"search"`
	Status   string `form:"status"`
	Channel  string `form:"channel"`
	Priority string `form:"priority"`
	Assignee *uint  `form:"assignee"`
	Limit    int    `form:"limit"`
	Offset   int    `form:"offset"`
}

// SupportTicketMessageResponse is a message in a ticket thread.
type SupportTicketMessageResponse struct {
	ID             uint      `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	TicketID       uint      `json:"ticket_id"`
	AuthorID       *uint     `json:"author_id,omitempty"`
	AuthorName     string    `json:"author_name,omitempty"`
	AuthorRole     string    `json:"author_role"`
	Body           string    `json:"body"`
	Channel        string    `json:"channel"`
	IsInternal     bool      `json:"is_internal"`
	IsAISuggestion bool      `json:"is_ai_suggestion"`
}

// SupportTicketResponse is the public/admin ticket projection.
type SupportTicketResponse struct {
	ID              uint                           `json:"id"`
	CreatedAt       time.Time                      `json:"created_at"`
	UpdatedAt       time.Time                      `json:"updated_at"`
	UserID          *uint                          `json:"user_id,omitempty"`
	CustomerName    string                         `json:"customer_name,omitempty"`
	CustomerEmail   string                         `json:"customer_email,omitempty"`
	OrderID         *uint                          `json:"order_id,omitempty"`
	OrderNumber     string                         `json:"order_number,omitempty"`
	Subject         string                         `json:"subject"`
	Status          string                         `json:"status"`
	Priority        string                         `json:"priority"`
	Channel         string                         `json:"channel"`
	AssigneeID      *uint                          `json:"assignee_id,omitempty"`
	AssigneeName    string                         `json:"assignee_name,omitempty"`
	AdminNotes      string                         `json:"admin_notes,omitempty"`
	LastMessageAt   *time.Time                     `json:"last_message_at,omitempty"`
	MessageCount    int64                          `json:"message_count,omitempty"`
	Messages        []SupportTicketMessageResponse `json:"messages,omitempty"`
}

// CreateSupportTicketMessageRequest adds a reply to a ticket.
type CreateSupportTicketMessageRequest struct {
	Body       string `json:"body" validate:"required,min=1,max=4096"`
	Channel    string `json:"channel" validate:"omitempty,oneof=email chat web internal"`
	IsInternal bool   `json:"is_internal"`
}

// UpdateSupportTicketNotesRequest updates internal admin notes.
type UpdateSupportTicketNotesRequest struct {
	AdminNotes string `json:"admin_notes" validate:"max=4096"`
}

// UpdateSupportTicketStatusRequest updates ticket workflow status.
type UpdateSupportTicketStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=open pending waiting_customer resolved closed"`
}

// UpdateSupportTicketAssigneeRequest assigns a staff member.
type UpdateSupportTicketAssigneeRequest struct {
	AssigneeID *uint `json:"assignee_id"`
}

// AdminSupportStats holds aggregate support desk metrics.
type AdminSupportStats struct {
	OpenTickets      int64 `json:"open_tickets"`
	PendingTickets   int64 `json:"pending_tickets"`
	ChatTickets      int64 `json:"chat_tickets"`
	EmailTickets     int64 `json:"email_tickets"`
	ResolvedToday    int64 `json:"resolved_today"`
	UnassignedTickets int64 `json:"unassigned_tickets"`
}

// SupportSuggestReplyResponse returns an AI-drafted reply for staff review.
type SupportSuggestReplyResponse struct {
	Suggestion string `json:"suggestion"`
}
