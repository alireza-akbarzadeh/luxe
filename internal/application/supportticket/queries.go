package supportticket

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Queries orchestrates support ticket read use cases.
type Queries struct {
	repo *postgres.SupportTicketRepository
}

// NewQueries creates support ticket query use cases.
func NewQueries(repo *postgres.SupportTicketRepository) *Queries {
	return &Queries{repo: repo}
}

// GetByID loads a ticket with messages.
func (q *Queries) GetByID(ctx context.Context, ticketID, userID uint, isAdmin bool) (dto.SupportTicketResponse, error) {
	ticket, err := q.repo.FindByID(ctx, ticketID, userID, isAdmin)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.SupportTicketResponse{}, utils.ErrNotFound("ticket not found")
		}
		return dto.SupportTicketResponse{}, utils.ErrInternal(err)
	}
	messages, err := q.repo.ListMessages(ctx, ticketID, isAdmin)
	if err != nil {
		return dto.SupportTicketResponse{}, utils.ErrInternal(err)
	}
	count, err := q.repo.CountMessages(ctx, ticketID, isAdmin)
	if err != nil {
		return dto.SupportTicketResponse{}, utils.ErrInternal(err)
	}
	return ToTicketResponse(ticket, messages, count), nil
}

// ListForUser returns paginated tickets for a user.
func (q *Queries) ListForUser(ctx context.Context, userID uint, limit, offset int) ([]dto.SupportTicketResponse, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	total, err := q.repo.CountForUser(ctx, userID)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	tickets, err := q.repo.ListForUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	items, err := q.mapTicketSummaries(ctx, tickets, false)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ListAdmin returns paginated tickets for admin.
func (q *Queries) ListAdmin(ctx context.Context, filters dto.AdminSupportTicketListFilters) ([]dto.SupportTicketResponse, int64, error) {
	limit, offset := filters.Limit, filters.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	total, err := q.repo.CountAdmin(ctx, filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	tickets, err := q.repo.ListAdmin(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	items, err := q.mapTicketSummaries(ctx, tickets, true)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetAdminStats loads aggregate support metrics.
func (q *Queries) GetAdminStats(ctx context.Context) (dto.AdminSupportStats, error) {
	stats, err := q.repo.AdminStats(ctx)
	if err != nil {
		return dto.AdminSupportStats{}, utils.ErrInternal(err)
	}
	return stats, nil
}

func (q *Queries) mapTicketSummaries(ctx context.Context, tickets []models.SupportTicket, includeInternal bool) ([]dto.SupportTicketResponse, error) {
	items := make([]dto.SupportTicketResponse, len(tickets))
	for i := range tickets {
		count, err := q.repo.CountMessages(ctx, tickets[i].ID, includeInternal)
		if err != nil {
			return nil, utils.ErrInternal(err)
		}
		items[i] = ToTicketResponse(&tickets[i], nil, count)
	}
	return items, nil
}
