package emailmarketing

import (
	"context"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service handles newsletter subscribers, templates, and email campaigns.
type Service struct {
	repo      *postgres.EmailMarketingRepository
	jobQueue  asynq.JobQueue
	frontendURL string
}

// NewService wires the email marketing application service.
func NewService(repo *postgres.EmailMarketingRepository, jobQueue asynq.JobQueue, frontendURL string) *Service {
	return &Service{repo: repo, jobQueue: jobQueue, frontendURL: strings.TrimRight(frontendURL, "/")}
}

func normalizeSubscriberStatus(status string) string {
	if strings.TrimSpace(strings.ToLower(status)) == "unsubscribed" {
		return "unsubscribed"
	}
	return "subscribed"
}

func normalizeTemplateStatus(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "active", "archived":
		return strings.TrimSpace(strings.ToLower(status))
	default:
		return "draft"
	}
}

func normalizeCampaignStatus(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "scheduled", "sending", "sent", "cancelled":
		return strings.TrimSpace(strings.ToLower(status))
	default:
		return "draft"
	}
}

func normalizeSegment(segment string) string {
	switch strings.TrimSpace(strings.ToLower(segment)) {
	case "checkout", "footer", "home", "register", "vip", "loyal", "new", "at_risk":
		return strings.TrimSpace(strings.ToLower(segment))
	default:
		return "all"
	}
}

func normalizeSource(source string) string {
	switch strings.TrimSpace(strings.ToLower(source)) {
	case "checkout", "footer", "home", "register", "manual":
		return strings.TrimSpace(strings.ToLower(source))
	default:
		return "home"
	}
}

func generateToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func slugify(value string) string {
	s := strings.ToLower(strings.TrimSpace(value))
	s = strings.ReplaceAll(s, " ", "-")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "template"
	}
	return out
}

// GetKPIs returns email marketing hub metrics.
func (s *Service) GetKPIs(ctx context.Context) (*dto.EmailMarketingKPIData, error) {
	active, err := s.repo.CountSubscribersByStatus(ctx, "subscribed")
	if err != nil {
		return nil, err
	}
	unsub, err := s.repo.CountSubscribersByStatus(ctx, "unsubscribed")
	if err != nil {
		return nil, err
	}
	templates, err := s.repo.CountTemplates(ctx)
	if err != nil {
		return nil, err
	}
	sentCampaigns, err := s.repo.CountCampaignsByStatus(ctx, "sent")
	if err != nil {
		return nil, err
	}
	scheduled, err := s.repo.CountCampaignsByStatus(ctx, "scheduled")
	if err != nil {
		return nil, err
	}
	delivered, err := s.repo.SumCampaignSentCount(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.EmailMarketingKPIData{
		TotalSubscribers:     active + unsub,
		ActiveSubscribers:    active,
		UnsubscribedCount:    unsub,
		TemplateCount:        templates,
		CampaignsSent:        sentCampaigns,
		CampaignsScheduled:   scheduled,
		TotalEmailsDelivered: delivered,
	}, nil
}

// Subscribe records a public newsletter opt-in.
func (s *Service) Subscribe(ctx context.Context, req dto.SubscribeNewsletterRequest, userID *uint) (*models.NewsletterSubscriber, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return nil, utils.ErrBadRequest("email is required")
	}

	token, err := generateToken()
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	row := &models.NewsletterSubscriber{
		Email:            email,
		UserID:           userID,
		Status:           "subscribed",
		Source:           normalizeSource(req.Source),
		UnsubscribeToken: token,
		SubscribedAt:     time.Now(),
	}
	if err := s.repo.UpsertSubscriber(ctx, row); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return s.repo.FindSubscriberByEmail(ctx, email)
}

// Unsubscribe marks a subscriber as unsubscribed by token.
func (s *Service) Unsubscribe(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return utils.ErrBadRequest("token is required")
	}
	if err := s.repo.UnsubscribeByToken(ctx, token); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// RecordCheckoutOptIn upserts a subscriber when checkout newsletter is checked.
func (s *Service) RecordCheckoutOptIn(ctx context.Context, email string, userID uint) {
	email = strings.TrimSpace(email)
	if email == "" {
		return
	}
	uid := userID
	_, _ = s.Subscribe(ctx, dto.SubscribeNewsletterRequest{Email: email, Source: "checkout"}, &uid)
}

// ListSubscribers returns paginated subscribers for admin.
func (s *Service) ListSubscribers(ctx context.Context, filters dto.AdminSubscriberListFilters) ([]models.NewsletterSubscriber, int64, error) {
	return s.repo.ListSubscribers(ctx, filters)
}

// ExportSubscribersCSV returns subscriber rows as CSV bytes.
func (s *Service) ExportSubscribersCSV(ctx context.Context, filters dto.AdminSubscriberListFilters) ([]byte, error) {
	rows, err := s.repo.ListSubscribersForExport(ctx, filters)
	if err != nil {
		return nil, err
	}

	var buf strings.Builder
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "email", "status", "source", "user_id", "subscribed_at", "unsubscribed_at"})
	for _, row := range rows {
		userID := ""
		if row.UserID != nil {
			userID = fmt.Sprintf("%d", *row.UserID)
		}
		unsub := ""
		if row.UnsubscribedAt != nil {
			unsub = row.UnsubscribedAt.Format(time.RFC3339)
		}
		_ = w.Write([]string{
			fmt.Sprintf("%d", row.ID),
			row.Email,
			row.Status,
			row.Source,
			userID,
			row.SubscribedAt.Format(time.RFC3339),
			unsub,
		})
	}
	w.Flush()
	return []byte(buf.String()), w.Error()
}

// DeleteSubscriber removes a subscriber record.
func (s *Service) DeleteSubscriber(ctx context.Context, id uint) error {
	if err := s.repo.DeleteSubscriber(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("subscriber not found")
		}
		return utils.ErrInternal(err)
	}
	return nil
}

// ListTemplates returns paginated email templates.
func (s *Service) ListTemplates(ctx context.Context, filters dto.AdminEmailTemplateListFilters) ([]models.EmailTemplate, int64, error) {
	return s.repo.ListTemplates(ctx, filters)
}

// GetTemplate returns a template by ID.
func (s *Service) GetTemplate(ctx context.Context, id uint) (*models.EmailTemplate, error) {
	row, err := s.repo.FindTemplateByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("template not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return row, nil
}

// CreateTemplate creates an email template.
func (s *Service) CreateTemplate(ctx context.Context, req dto.CreateEmailTemplateRequest) (*models.EmailTemplate, error) {
	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = slugify(req.Name)
	}
	row := &models.EmailTemplate{
		Name:     strings.TrimSpace(req.Name),
		Slug:     slug,
		Subject:  strings.TrimSpace(req.Subject),
		BodyHTML: req.BodyHTML,
		Status:   normalizeTemplateStatus(req.Status),
	}
	if err := s.repo.CreateTemplate(ctx, row); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return row, nil
}

// UpdateTemplate updates an email template.
func (s *Service) UpdateTemplate(ctx context.Context, id uint, req dto.UpdateEmailTemplateRequest) (*models.EmailTemplate, error) {
	row, err := s.GetTemplate(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		row.Name = strings.TrimSpace(*req.Name)
	}
	if req.Slug != nil {
		row.Slug = strings.TrimSpace(*req.Slug)
	}
	if req.Subject != nil {
		row.Subject = strings.TrimSpace(*req.Subject)
	}
	if req.BodyHTML != nil {
		row.BodyHTML = *req.BodyHTML
	}
	if req.Status != nil {
		row.Status = normalizeTemplateStatus(*req.Status)
	}
	if err := s.repo.UpdateTemplate(ctx, row); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return row, nil
}

// DeleteTemplate removes a template.
func (s *Service) DeleteTemplate(ctx context.Context, id uint) error {
	if err := s.repo.DeleteTemplate(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("template not found")
		}
		return utils.ErrInternal(err)
	}
	return nil
}

// ListCampaigns returns paginated email campaigns.
func (s *Service) ListCampaigns(ctx context.Context, filters dto.AdminEmailCampaignListFilters) ([]models.EmailCampaign, int64, error) {
	return s.repo.ListCampaigns(ctx, filters)
}

// GetCampaign returns a campaign by ID.
func (s *Service) GetCampaign(ctx context.Context, id uint) (*models.EmailCampaign, error) {
	row, err := s.repo.FindCampaignByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("campaign not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return row, nil
}

// CreateCampaign creates an email campaign.
func (s *Service) CreateCampaign(ctx context.Context, req dto.CreateEmailCampaignRequest) (*models.EmailCampaign, error) {
	row := &models.EmailCampaign{
		Name:        strings.TrimSpace(req.Name),
		Subject:     strings.TrimSpace(req.Subject),
		BodyHTML:    req.BodyHTML,
		TemplateID:  req.TemplateID,
		Segment:     normalizeSegment(req.Segment),
		Status:      normalizeCampaignStatus(req.Status),
		ScheduledAt: req.ScheduledAt,
	}
	if row.Status == "scheduled" && row.ScheduledAt == nil {
		return nil, utils.ErrBadRequest("scheduled_at is required for scheduled campaigns")
	}
	if err := s.repo.CreateCampaign(ctx, row); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return s.repo.FindCampaignByID(ctx, row.ID)
}

// UpdateCampaign updates an email campaign.
func (s *Service) UpdateCampaign(ctx context.Context, id uint, req dto.UpdateEmailCampaignRequest) (*models.EmailCampaign, error) {
	row, err := s.GetCampaign(ctx, id)
	if err != nil {
		return nil, err
	}
	if row.Status == "sent" || row.Status == "sending" {
		return nil, utils.ErrBadRequest("sent campaigns cannot be edited")
	}
	if req.Name != nil {
		row.Name = strings.TrimSpace(*req.Name)
	}
	if req.Subject != nil {
		row.Subject = strings.TrimSpace(*req.Subject)
	}
	if req.BodyHTML != nil {
		row.BodyHTML = *req.BodyHTML
	}
	if req.TemplateID != nil {
		row.TemplateID = req.TemplateID
	}
	if req.Segment != nil {
		row.Segment = normalizeSegment(*req.Segment)
	}
	if req.Status != nil {
		row.Status = normalizeCampaignStatus(*req.Status)
	}
	if req.ScheduledAt != nil {
		row.ScheduledAt = req.ScheduledAt
	}
	if row.Status == "scheduled" && row.ScheduledAt == nil {
		return nil, utils.ErrBadRequest("scheduled_at is required for scheduled campaigns")
	}
	if err := s.repo.UpdateCampaign(ctx, row); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return row, nil
}

// DeleteCampaign removes a draft/scheduled campaign.
func (s *Service) DeleteCampaign(ctx context.Context, id uint) error {
	row, err := s.GetCampaign(ctx, id)
	if err != nil {
		return err
	}
	if row.Status == "sending" || row.Status == "sent" {
		return utils.ErrBadRequest("cannot delete a campaign that has been sent")
	}
	if err := s.repo.DeleteCampaign(ctx, id); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// ScheduleCampaign sets a campaign to scheduled status.
func (s *Service) ScheduleCampaign(ctx context.Context, id uint, scheduledAt time.Time) (*models.EmailCampaign, error) {
	row, err := s.GetCampaign(ctx, id)
	if err != nil {
		return nil, err
	}
	if row.Status == "sent" || row.Status == "sending" {
		return nil, utils.ErrBadRequest("campaign already sent")
	}
	if scheduledAt.Before(time.Now()) {
		return nil, utils.ErrBadRequest("scheduled_at must be in the future")
	}
	row.Status = "scheduled"
	row.ScheduledAt = &scheduledAt
	if err := s.repo.UpdateCampaign(ctx, row); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return row, nil
}

func (s *Service) resolveCampaignContent(ctx context.Context, campaign *models.EmailCampaign) (subject, body string, err error) {
	subject = campaign.Subject
	body = campaign.BodyHTML
	if campaign.TemplateID != nil && *campaign.TemplateID > 0 {
		tpl, err := s.repo.FindTemplateByID(ctx, *campaign.TemplateID)
		if err != nil {
			return "", "", utils.ErrBadRequest("linked template not found")
		}
		if strings.TrimSpace(subject) == "" {
			subject = tpl.Subject
		}
		if strings.TrimSpace(body) == "" {
			body = tpl.BodyHTML
		}
	}
	if strings.TrimSpace(subject) == "" || strings.TrimSpace(body) == "" {
		return "", "", utils.ErrBadRequest("campaign subject and body are required")
	}
	return subject, body, nil
}

func appendUnsubscribeFooter(body, token, frontendURL string) string {
	if token == "" || frontendURL == "" {
		return body
	}
	link := fmt.Sprintf("%s/newsletters/unsubscribe?token=%s", frontendURL, token)
	return body + fmt.Sprintf(`<p style="margin-top:24px;font-size:12px;color:#666;">`+
		`<a href="%s">Unsubscribe</a> from marketing emails.</p>`, link)
}

// SendCampaign enqueues emails to the campaign segment via the job queue.
func (s *Service) SendCampaign(ctx context.Context, id uint) (*dto.EmailCampaignSendData, error) {
	campaign, err := s.GetCampaign(ctx, id)
	if err != nil {
		return nil, err
	}
	if campaign.Status == "sent" || campaign.Status == "sending" {
		return nil, utils.ErrBadRequest("campaign already sent")
	}

	subject, body, err := s.resolveCampaignContent(ctx, campaign)
	if err != nil {
		return nil, err
	}

	subscribers, err := s.repo.ListSubscribersBySegment(ctx, campaign.Segment)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	if len(subscribers) == 0 {
		return nil, utils.ErrBadRequest("no subscribers match this segment")
	}

	campaign.Status = "sending"
	campaign.RecipientCount = len(subscribers)
	campaign.SentCount = 0
	campaign.FailedCount = 0
	if err := s.repo.UpdateCampaign(ctx, campaign); err != nil {
		return nil, utils.ErrInternal(err)
	}

	enqueued := 0
	failed := 0
	for _, sub := range subscribers {
		personalBody := appendUnsubscribeFooter(body, sub.UnsubscribeToken, s.frontendURL)
		if err := s.jobQueue.EnqueueSendEmail(ctx, sub.Email, subject, personalBody); err != nil {
			failed++
			continue
		}
		enqueued++
	}

	now := time.Now()
	campaign.SentCount = enqueued
	campaign.FailedCount = failed
	campaign.Status = "sent"
	campaign.SentAt = &now
	if err := s.repo.UpdateCampaign(ctx, campaign); err != nil {
		return nil, utils.ErrInternal(err)
	}

	return &dto.EmailCampaignSendData{
		CampaignID:     campaign.ID,
		RecipientCount: campaign.RecipientCount,
		EnqueuedCount:  enqueued,
		FailedCount:    failed,
	}, nil
}
