package supportticket

import (
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

func authorDisplayName(user *models.User, fallbackRole string) string {
	if user == nil {
		return fallbackRole
	}
	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if name != "" {
		return name
	}
	if user.Email != "" {
		return user.Email
	}
	return fallbackRole
}

// ToMessageResponse maps a message model to API shape.
func ToMessageResponse(msg *models.SupportTicketMessage) dto.SupportTicketMessageResponse {
	if msg == nil {
		return dto.SupportTicketMessageResponse{}
	}
	name := authorDisplayName(msg.Author, msg.AuthorRole)
	return dto.SupportTicketMessageResponse{
		ID:             msg.ID,
		CreatedAt:      msg.CreatedAt,
		TicketID:       msg.TicketID,
		AuthorID:       msg.AuthorID,
		AuthorName:     name,
		AuthorRole:     msg.AuthorRole,
		Body:           msg.Body,
		Channel:        msg.Channel,
		IsInternal:     msg.IsInternal,
		IsAISuggestion: msg.IsAISuggestion,
	}
}

// ToTicketResponse maps a ticket model to API shape.
func ToTicketResponse(ticket *models.SupportTicket, messages []models.SupportTicketMessage, messageCount int64) dto.SupportTicketResponse {
	if ticket == nil {
		return dto.SupportTicketResponse{}
	}
	resp := dto.SupportTicketResponse{
		ID:            ticket.ID,
		CreatedAt:     ticket.CreatedAt,
		UpdatedAt:     ticket.UpdatedAt,
		UserID:        ticket.UserID,
		CustomerName:  ticket.CustomerName,
		CustomerEmail: ticket.CustomerEmail,
		OrderID:       ticket.OrderID,
		Subject:       ticket.Subject,
		Status:        ticket.Status,
		Priority:      ticket.Priority,
		Channel:       ticket.Channel,
		AssigneeID:    ticket.AssigneeID,
		AdminNotes:    ticket.AdminNotes,
		LastMessageAt: ticket.LastMessageAt,
		MessageCount:  messageCount,
	}
	if ticket.User != nil {
		if resp.CustomerName == "" {
			resp.CustomerName = authorDisplayName(ticket.User, "Customer")
		}
		if resp.CustomerEmail == "" {
			resp.CustomerEmail = ticket.User.Email
		}
	}
	if ticket.Assignee != nil {
		resp.AssigneeName = authorDisplayName(ticket.Assignee, "Staff")
	}
	if ticket.Order != nil {
		resp.OrderNumber = ticket.Order.OrderNumber
	}
	if len(messages) > 0 {
		resp.Messages = make([]dto.SupportTicketMessageResponse, len(messages))
		for i := range messages {
			resp.Messages[i] = ToMessageResponse(&messages[i])
		}
	}
	return resp
}
