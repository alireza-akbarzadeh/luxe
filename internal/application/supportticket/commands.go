package supportticket

import (
	"context"
	"errors"
	"fmt"
	"strings"

	appai "github.com/alireza-akbarzadeh/luxe/internal/application/ai"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Commands orchestrates support ticket write use cases.
type Commands struct {
	repo *postgres.SupportTicketRepository
	ai   *appai.Service
}

// NewCommands creates support ticket command use cases.
func NewCommands(repo *postgres.SupportTicketRepository, ai *appai.Service) *Commands {
	return &Commands{repo: repo, ai: ai}
}

// Create opens a ticket from a customer (authenticated or guest).
func (c *Commands) Create(ctx context.Context, userID *uint, userEmail, userName string, req dto.CreateSupportTicketRequest) (*models.SupportTicket, error) {
	channel := strings.TrimSpace(req.Channel)
	if channel == "" {
		channel = constants.SupportChannelWeb
	}
	priority := strings.TrimSpace(req.Priority)
	if priority == "" {
		priority = constants.SupportPriorityNormal
	}

	email := strings.TrimSpace(req.CustomerEmail)
	name := strings.TrimSpace(req.CustomerName)
	if userID != nil && *userID > 0 {
		if email == "" {
			email = userEmail
		}
		if name == "" {
			name = userName
		}
	}
	if email == "" {
		return nil, utils.ErrBadRequest("customer_email is required")
	}

	ticket := &models.SupportTicket{
		UserID:        userID,
		CustomerName:  name,
		CustomerEmail: email,
		OrderID:       req.OrderID,
		Subject:       strings.TrimSpace(req.Subject),
		Status:        constants.SupportTicketStatusOpen,
		Priority:      priority,
		Channel:       channel,
	}
	firstMessage := &models.SupportTicketMessage{
		AuthorID:   userID,
		AuthorRole: constants.SupportAuthorCustomer,
		Body:       strings.TrimSpace(req.Message),
		Channel:    channel,
	}
	if err := c.repo.CreateTicket(ctx, ticket, firstMessage); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return ticket, nil
}

// AddMessage appends a reply to a ticket.
func (c *Commands) AddMessage(ctx context.Context, ticketID uint, authorID *uint, authorRole string, req dto.CreateSupportTicketMessageRequest, isAdmin bool) (*models.SupportTicketMessage, error) {
	ticket, err := c.loadTicket(ctx, ticketID, authorID, isAdmin)
	if err != nil {
		return nil, err
	}

	channel := strings.TrimSpace(req.Channel)
	if channel == "" {
		channel = ticket.Channel
	}
	if channel == "internal" {
		req.IsInternal = true
	}
	if req.IsInternal && !isAdmin {
		return nil, utils.ErrForbidden("internal notes require staff access")
	}
	if !isAdmin && ticket.UserID != nil && authorID != nil && *ticket.UserID != *authorID {
		return nil, utils.ErrForbidden("not allowed to reply on this ticket")
	}

	msg := &models.SupportTicketMessage{
		TicketID:   ticket.ID,
		AuthorID:   authorID,
		AuthorRole: authorRole,
		Body:       strings.TrimSpace(req.Body),
		Channel:    channel,
		IsInternal: req.IsInternal,
	}
	if err := c.repo.AddMessage(ctx, msg); err != nil {
		return nil, utils.ErrInternal(err)
	}

	if isAdmin && ticket.Status == constants.SupportTicketStatusOpen {
		_, _ = c.repo.UpdateStatus(ctx, ticket.ID, constants.SupportTicketStatusPending)
	}
	if !isAdmin && ticket.Status != constants.SupportTicketStatusClosed {
		_, _ = c.repo.UpdateStatus(ctx, ticket.ID, constants.SupportTicketStatusWaitingCustomer)
	}

	return msg, nil
}

// UpdateNotes sets internal admin notes.
func (c *Commands) UpdateNotes(ctx context.Context, ticketID uint, notes string) error {
	rows, err := c.repo.UpdateAdminNotes(ctx, ticketID, notes)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("ticket not found")
	}
	return nil
}

// UpdateStatus sets ticket status.
func (c *Commands) UpdateStatus(ctx context.Context, ticketID uint, status string) error {
	rows, err := c.repo.UpdateStatus(ctx, ticketID, status)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("ticket not found")
	}
	return nil
}

// UpdateAssignee sets ticket assignee.
func (c *Commands) UpdateAssignee(ctx context.Context, ticketID uint, assigneeID *uint) error {
	rows, err := c.repo.UpdateAssignee(ctx, ticketID, assigneeID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("ticket not found")
	}
	return nil
}

// SuggestReply drafts an AI reply for staff review.
func (c *Commands) SuggestReply(ctx context.Context, staffID, ticketID uint) (string, error) {
	if c.ai == nil {
		return "", utils.NewAppError(503, "AI is not available", nil)
	}
	ticket, err := c.repo.FindByID(ctx, ticketID, 0, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", utils.ErrNotFound("ticket not found")
		}
		return "", utils.ErrInternal(err)
	}
	messages, err := c.repo.ListMessages(ctx, ticketID, true)
	if err != nil {
		return "", utils.ErrInternal(err)
	}

	var transcript strings.Builder
	for _, msg := range messages {
		if msg.IsInternal {
			continue
		}
		fmt.Fprintf(&transcript, "%s: %s\n", msg.AuthorRole, msg.Body)
	}

	resp, err := c.ai.Generate(ctx, staffID, dto.AiGenerateRequest{
		Task: appai.TaskSupportReply,
		Context: map[string]any{
			"subject":    ticket.Subject,
			"channel":    ticket.Channel,
			"status":     ticket.Status,
			"customer":   ticket.CustomerName,
			"transcript": transcript.String(),
		},
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp.Text), nil
}

func (c *Commands) loadTicket(ctx context.Context, ticketID uint, userID *uint, isAdmin bool) (*models.SupportTicket, error) {
	uid := uint(0)
	if userID != nil {
		uid = *userID
	}
	ticket, err := c.repo.FindByID(ctx, ticketID, uid, isAdmin)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("ticket not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return ticket, nil
}
