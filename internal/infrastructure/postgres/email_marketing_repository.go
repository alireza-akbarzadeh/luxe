package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// EmailMarketingRepository handles newsletter subscribers, templates, and campaigns.
type EmailMarketingRepository struct {
	db *gorm.DB
}

// NewEmailMarketingRepository creates an email marketing repository.
func NewEmailMarketingRepository(db *gorm.DB) *EmailMarketingRepository {
	return &EmailMarketingRepository{db: db}
}

func listLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// --- Subscribers ---

func (r *EmailMarketingRepository) ListSubscribers(ctx context.Context, filters dto.AdminSubscriberListFilters) ([]models.NewsletterSubscriber, int64, error) {
	limit, offset := listLimitOffset(filters.Limit, filters.Offset)
	query := r.db.WithContext(ctx).Model(&models.NewsletterSubscriber{}).Preload("User")

	if filters.Status != "" && filters.Status != "all" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Source != "" && filters.Source != "all" {
		query = query.Where("source = ?", filters.Source)
	}
	if filters.Search != "" {
		like := "%" + strings.TrimSpace(filters.Search) + "%"
		query = query.Where("email ILIKE ?", like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []models.NewsletterSubscriber
	err := query.Order("subscribed_at DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}

func (r *EmailMarketingRepository) ListSubscribersForExport(ctx context.Context, filters dto.AdminSubscriberListFilters) ([]models.NewsletterSubscriber, error) {
	query := r.db.WithContext(ctx).Model(&models.NewsletterSubscriber{})
	if filters.Status != "" && filters.Status != "all" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Source != "" && filters.Source != "all" {
		query = query.Where("source = ?", filters.Source)
	}
	if filters.Search != "" {
		like := "%" + strings.TrimSpace(filters.Search) + "%"
		query = query.Where("email ILIKE ?", like)
	}
	var rows []models.NewsletterSubscriber
	err := query.Order("subscribed_at DESC").Limit(10000).Find(&rows).Error
	return rows, err
}

func (r *EmailMarketingRepository) FindSubscriberByID(ctx context.Context, id uint) (*models.NewsletterSubscriber, error) {
	var row models.NewsletterSubscriber
	err := r.db.WithContext(ctx).Preload("User").First(&row, id).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *EmailMarketingRepository) FindSubscriberByEmail(ctx context.Context, email string) (*models.NewsletterSubscriber, error) {
	var row models.NewsletterSubscriber
	err := r.db.WithContext(ctx).
		Where("LOWER(email) = LOWER(?)", strings.TrimSpace(email)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *EmailMarketingRepository) FindSubscriberByToken(ctx context.Context, token string) (*models.NewsletterSubscriber, error) {
	var row models.NewsletterSubscriber
	err := r.db.WithContext(ctx).Where("unsubscribe_token = ?", token).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *EmailMarketingRepository) UpsertSubscriber(ctx context.Context, row *models.NewsletterSubscriber) error {
	existing, err := r.FindSubscriberByEmail(ctx, row.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(row).Error
	}

	existing.Status = "subscribed"
	existing.Source = row.Source
	if row.UserID != nil {
		existing.UserID = row.UserID
	}
	existing.UnsubscribedAt = nil
	existing.SubscribedAt = time.Now()
	return r.db.WithContext(ctx).Save(existing).Error
}

func (r *EmailMarketingRepository) UnsubscribeByToken(ctx context.Context, token string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.NewsletterSubscriber{}).
		Where("unsubscribe_token = ? AND status = ?", token, "subscribed").
		Updates(map[string]interface{}{
			"status":          "unsubscribed",
			"unsubscribed_at": now,
		}).Error
}

func (r *EmailMarketingRepository) DeleteSubscriber(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.NewsletterSubscriber{}, id).Error
}

func (r *EmailMarketingRepository) CountSubscribersByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.NewsletterSubscriber{}).
		Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *EmailMarketingRepository) ListSubscribersBySegment(ctx context.Context, segment string) ([]models.NewsletterSubscriber, error) {
	query := r.db.WithContext(ctx).Model(&models.NewsletterSubscriber{}).
		Where("status = ?", "subscribed")

	switch strings.TrimSpace(strings.ToLower(segment)) {
	case "checkout":
		query = query.Where("source = ?", "checkout")
	case "footer", "home":
		query = query.Where("source IN ?", []string{"footer", "home"})
	case "register":
		query = query.Where("source = ?", "register")
	case "vip", "loyal", "new", "at_risk":
		query = query.Joins("JOIN users ON users.id = newsletter_subscribers.user_id").
			Where("users.customer_segment = ?", segment)
	default:
		// all subscribed
	}

	var rows []models.NewsletterSubscriber
	err := query.Order("id ASC").Limit(50000).Find(&rows).Error
	return rows, err
}

// --- Templates ---

func (r *EmailMarketingRepository) ListTemplates(ctx context.Context, filters dto.AdminEmailTemplateListFilters) ([]models.EmailTemplate, int64, error) {
	limit, offset := listLimitOffset(filters.Limit, filters.Offset)
	query := r.db.WithContext(ctx).Model(&models.EmailTemplate{})

	if filters.Status != "" && filters.Status != "all" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Search != "" {
		like := "%" + strings.TrimSpace(filters.Search) + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ? OR subject ILIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []models.EmailTemplate
	err := query.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}

func (r *EmailMarketingRepository) FindTemplateByID(ctx context.Context, id uint) (*models.EmailTemplate, error) {
	var row models.EmailTemplate
	err := r.db.WithContext(ctx).First(&row, id).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *EmailMarketingRepository) CreateTemplate(ctx context.Context, row *models.EmailTemplate) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *EmailMarketingRepository) UpdateTemplate(ctx context.Context, row *models.EmailTemplate) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *EmailMarketingRepository) DeleteTemplate(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.EmailTemplate{}, id).Error
}

func (r *EmailMarketingRepository) CountTemplates(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.EmailTemplate{}).Count(&count).Error
	return count, err
}

// --- Campaigns ---

func (r *EmailMarketingRepository) ListCampaigns(ctx context.Context, filters dto.AdminEmailCampaignListFilters) ([]models.EmailCampaign, int64, error) {
	limit, offset := listLimitOffset(filters.Limit, filters.Offset)
	query := r.db.WithContext(ctx).Model(&models.EmailCampaign{}).Preload("Template")

	if filters.Status != "" && filters.Status != "all" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Search != "" {
		like := "%" + strings.TrimSpace(filters.Search) + "%"
		query = query.Where("name ILIKE ? OR subject ILIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []models.EmailCampaign
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}

func (r *EmailMarketingRepository) FindCampaignByID(ctx context.Context, id uint) (*models.EmailCampaign, error) {
	var row models.EmailCampaign
	err := r.db.WithContext(ctx).Preload("Template").First(&row, id).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *EmailMarketingRepository) CreateCampaign(ctx context.Context, row *models.EmailCampaign) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *EmailMarketingRepository) UpdateCampaign(ctx context.Context, row *models.EmailCampaign) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *EmailMarketingRepository) DeleteCampaign(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.EmailCampaign{}, id).Error
}

func (r *EmailMarketingRepository) CountCampaignsByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.EmailCampaign{}).
		Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *EmailMarketingRepository) SumCampaignSentCount(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.EmailCampaign{}).
		Select("COALESCE(SUM(sent_count), 0)").Scan(&total).Error
	return total, err
}
