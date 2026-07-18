package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// SupportTicketRepository persists support tickets and messages.
type SupportTicketRepository struct {
	db *gorm.DB
}

// NewSupportTicketRepository creates a GORM-backed support repository.
func NewSupportTicketRepository(db *gorm.DB) *SupportTicketRepository {
	return &SupportTicketRepository{db: db}
}

// CreateTicket inserts a ticket and its first customer message atomically.
func (r *SupportTicketRepository) CreateTicket(ctx context.Context, ticket *models.SupportTicket, firstMessage *models.SupportTicketMessage) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ticket).Error; err != nil {
			return err
		}
		firstMessage.TicketID = ticket.ID
		if err := tx.Create(firstMessage).Error; err != nil {
			return err
		}
		now := firstMessage.CreatedAt
		return tx.Model(ticket).Update("last_message_at", now).Error
	})
}

// AddMessage appends a message and bumps last_message_at.
func (r *SupportTicketRepository) AddMessage(ctx context.Context, msg *models.SupportTicketMessage) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(msg).Error; err != nil {
			return err
		}
		return tx.Model(&models.SupportTicket{}).
			Where("id = ?", msg.TicketID).
			Updates(map[string]any{
				"last_message_at": msg.CreatedAt,
				"updated_at":      time.Now(),
			}).Error
	})
}

// FindByID loads a ticket with optional user scope.
func (r *SupportTicketRepository) FindByID(ctx context.Context, ticketID uint, userID uint, isAdmin bool) (*models.SupportTicket, error) {
	q := r.db.WithContext(ctx).
		Preload("User").
		Preload("Assignee").
		Preload("Order").
		Where("id = ?", ticketID)
	if !isAdmin {
		q = q.Where("user_id = ?", userID)
	}
	var ticket models.SupportTicket
	if err := q.First(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

// ListMessages returns messages for a ticket; non-admin views hide internal notes.
func (r *SupportTicketRepository) ListMessages(ctx context.Context, ticketID uint, includeInternal bool) ([]models.SupportTicketMessage, error) {
	q := r.db.WithContext(ctx).
		Preload("Author").
		Where("ticket_id = ?", ticketID).
		Order("created_at ASC")
	if !includeInternal {
		q = q.Where("is_internal = ?", false)
	}
	var messages []models.SupportTicketMessage
	err := q.Find(&messages).Error
	return messages, err
}

func (r *SupportTicketRepository) applyAdminFilters(q *gorm.DB, filters dto.AdminSupportTicketListFilters) *gorm.DB {
	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	}
	if filters.Channel != "" {
		q = q.Where("channel = ?", filters.Channel)
	}
	if filters.Priority != "" {
		q = q.Where("priority = ?", filters.Priority)
	}
	if filters.Assignee != nil {
		q = q.Where("assignee_id = ?", *filters.Assignee)
	}
	if filters.Search != "" {
		term := "%" + strings.ToLower(filters.Search) + "%"
		q = q.Where(
			"LOWER(subject) LIKE ? OR LOWER(customer_email) LIKE ? OR LOWER(customer_name) LIKE ?",
			term, term, term,
		)
	}
	return q
}

// CountAdmin counts tickets matching admin filters.
func (r *SupportTicketRepository) CountAdmin(ctx context.Context, filters dto.AdminSupportTicketListFilters) (int64, error) {
	q := r.applyAdminFilters(r.db.WithContext(ctx).Model(&models.SupportTicket{}), filters)
	var total int64
	err := q.Count(&total).Error
	return total, err
}

// ListAdmin returns paginated tickets for admin.
func (r *SupportTicketRepository) ListAdmin(ctx context.Context, filters dto.AdminSupportTicketListFilters, limit, offset int) ([]models.SupportTicket, error) {
	q := r.applyAdminFilters(r.db.WithContext(ctx).Model(&models.SupportTicket{}), filters)
	var tickets []models.SupportTicket
	err := q.Preload("User").Preload("Assignee").Preload("Order").
		Order("COALESCE(last_message_at, created_at) DESC").
		Limit(limit).Offset(offset).
		Find(&tickets).Error
	return tickets, err
}

// CountForUser counts tickets for a user.
func (r *SupportTicketRepository) CountForUser(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.SupportTicket{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	return total, err
}

// ListForUser returns paginated tickets for a user.
func (r *SupportTicketRepository) ListForUser(ctx context.Context, userID uint, limit, offset int) ([]models.SupportTicket, error) {
	var tickets []models.SupportTicket
	err := r.db.WithContext(ctx).Model(&models.SupportTicket{}).
		Where("user_id = ?", userID).
		Order("COALESCE(last_message_at, created_at) DESC").
		Limit(limit).Offset(offset).
		Find(&tickets).Error
	return tickets, err
}

// UpdateAdminNotes sets internal notes on a ticket.
func (r *SupportTicketRepository) UpdateAdminNotes(ctx context.Context, ticketID uint, notes string) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.SupportTicket{}).
		Where("id = ?", ticketID).
		Update("admin_notes", notes)
	return result.RowsAffected, result.Error
}

// UpdateStatus sets ticket status.
func (r *SupportTicketRepository) UpdateStatus(ctx context.Context, ticketID uint, status string) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.SupportTicket{}).
		Where("id = ?", ticketID).
		Update("status", status)
	return result.RowsAffected, result.Error
}

// UpdateAssignee sets ticket assignee.
func (r *SupportTicketRepository) UpdateAssignee(ctx context.Context, ticketID uint, assigneeID *uint) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.SupportTicket{}).
		Where("id = ?", ticketID).
		Update("assignee_id", assigneeID)
	return result.RowsAffected, result.Error
}

// CountMessages counts visible messages for a ticket.
func (r *SupportTicketRepository) CountMessages(ctx context.Context, ticketID uint, includeInternal bool) (int64, error) {
	q := r.db.WithContext(ctx).Model(&models.SupportTicketMessage{}).
		Where("ticket_id = ?", ticketID)
	if !includeInternal {
		q = q.Where("is_internal = ?", false)
	}
	var count int64
	err := q.Count(&count).Error
	return count, err
}

// AdminStats loads aggregate support metrics.
func (r *SupportTicketRepository) AdminStats(ctx context.Context) (dto.AdminSupportStats, error) {
	var stats dto.AdminSupportStats
	base := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&models.SupportTicket{})
	}
	terminal := []string{constants.SupportTicketStatusClosed, constants.SupportTicketStatusResolved}

	if err := base().Where("status IN ?", []string{constants.SupportTicketStatusOpen, constants.SupportTicketStatusPending}).
		Count(&stats.OpenTickets).Error; err != nil {
		return stats, err
	}
	if err := base().Where("status = ?", constants.SupportTicketStatusPending).Count(&stats.PendingTickets).Error; err != nil {
		return stats, err
	}
	if err := base().Where("channel = ? AND status NOT IN ?", constants.SupportChannelChat, terminal).
		Count(&stats.ChatTickets).Error; err != nil {
		return stats, err
	}
	if err := base().Where("channel = ? AND status NOT IN ?", constants.SupportChannelEmail, terminal).
		Count(&stats.EmailTickets).Error; err != nil {
		return stats, err
	}
	startOfDay := time.Now().UTC().Truncate(24 * time.Hour)
	if err := base().Where("status IN ? AND updated_at >= ?", terminal, startOfDay).
		Count(&stats.ResolvedToday).Error; err != nil {
		return stats, err
	}
	if err := base().Where("assignee_id IS NULL AND status NOT IN ?", terminal).
		Count(&stats.UnassignedTickets).Error; err != nil {
		return stats, err
	}
	return stats, nil
}
