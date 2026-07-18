package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// --- KPIs ---

type EmailMarketingKPIData struct {
	TotalSubscribers     int64 `json:"total_subscribers"`
	ActiveSubscribers    int64 `json:"active_subscribers"`
	UnsubscribedCount    int64 `json:"unsubscribed_count"`
	TemplateCount        int64 `json:"template_count"`
	CampaignsSent        int64 `json:"campaigns_sent"`
	CampaignsScheduled   int64 `json:"campaigns_scheduled"`
	TotalEmailsDelivered int64 `json:"total_emails_delivered"`
}

type EmailMarketingKPIResponse struct {
	BaseResponse
	Data EmailMarketingKPIData `json:"data"`
}

// --- Subscribers ---

type AdminSubscriberListFilters struct {
	Limit  int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset int    `form:"offset" validate:"omitempty,min=0"`
	Status string `form:"status" validate:"omitempty,oneof=all subscribed unsubscribed"`
	Source string `form:"source" validate:"omitempty,oneof=all checkout footer home register manual"`
	Search string `form:"search" validate:"omitempty"`
}

type SubscribeNewsletterRequest struct {
	Email  string `json:"email" validate:"required,email,max=255"`
	Source string `json:"source" validate:"omitempty,oneof=footer home checkout register"`
}

type UnsubscribeNewsletterRequest struct {
	Token string `json:"token" validate:"required,min=8,max=64"`
}

type SubscriberListResponse struct {
	BaseResponse
	Data SubscriberListData `json:"data"`
}

type SubscriberListData struct {
	Subscribers []models.NewsletterSubscriber `json:"subscribers"`
	Total       int64                         `json:"total"`
	Limit       int                           `json:"limit"`
	Offset      int                           `json:"offset"`
}

type SubscriberSingleResponse struct {
	BaseResponse
	Data SubscriberData `json:"data"`
}

type SubscriberData struct {
	Subscriber models.NewsletterSubscriber `json:"subscriber"`
}

// --- Templates ---

type AdminEmailTemplateListFilters struct {
	Limit  int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset int    `form:"offset" validate:"omitempty,min=0"`
	Status string `form:"status" validate:"omitempty,oneof=all draft active archived"`
	Search string `form:"search" validate:"omitempty"`
}

type CreateEmailTemplateRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=255"`
	Slug     string `json:"slug" validate:"required,min=2,max=128"`
	Subject  string `json:"subject" validate:"required,min=2,max=512"`
	BodyHTML string `json:"body_html"`
	Status   string `json:"status" validate:"omitempty,oneof=draft active archived"`
}

type UpdateEmailTemplateRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Slug     *string `json:"slug,omitempty" validate:"omitempty,min=2,max=128"`
	Subject  *string `json:"subject,omitempty" validate:"omitempty,min=2,max=512"`
	BodyHTML *string `json:"body_html,omitempty"`
	Status   *string `json:"status,omitempty" validate:"omitempty,oneof=draft active archived"`
}

type EmailTemplateListResponse struct {
	BaseResponse
	Data EmailTemplateListData `json:"data"`
}

type EmailTemplateListData struct {
	Templates []models.EmailTemplate `json:"templates"`
	Total     int64                  `json:"total"`
	Limit     int                    `json:"limit"`
	Offset    int                    `json:"offset"`
}

type EmailTemplateSingleResponse struct {
	BaseResponse
	Data EmailTemplateData `json:"data"`
}

type EmailTemplateData struct {
	Template models.EmailTemplate `json:"template"`
}

// --- Campaigns ---

type AdminEmailCampaignListFilters struct {
	Limit  int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset int    `form:"offset" validate:"omitempty,min=0"`
	Status string `form:"status" validate:"omitempty,oneof=all draft scheduled sending sent cancelled"`
	Search string `form:"search" validate:"omitempty"`
}

type CreateEmailCampaignRequest struct {
	Name        string     `json:"name" validate:"required,min=2,max=255"`
	Subject     string     `json:"subject" validate:"required,min=2,max=512"`
	BodyHTML    string     `json:"body_html"`
	TemplateID  *uint      `json:"template_id,omitempty" validate:"omitempty,gt=0"`
	Segment     string     `json:"segment" validate:"omitempty,oneof=all checkout footer home register vip loyal new at_risk"`
	Status      string     `json:"status" validate:"omitempty,oneof=draft scheduled"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

type UpdateEmailCampaignRequest struct {
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Subject     *string    `json:"subject,omitempty" validate:"omitempty,min=2,max=512"`
	BodyHTML    *string    `json:"body_html,omitempty"`
	TemplateID  *uint      `json:"template_id,omitempty" validate:"omitempty,gt=0"`
	Segment     *string    `json:"segment,omitempty" validate:"omitempty,oneof=all checkout footer home register vip loyal new at_risk"`
	Status      *string    `json:"status,omitempty" validate:"omitempty,oneof=draft scheduled cancelled"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

type ScheduleEmailCampaignRequest struct {
	ScheduledAt time.Time `json:"scheduled_at" validate:"required"`
}

type EmailCampaignListResponse struct {
	BaseResponse
	Data EmailCampaignListData `json:"data"`
}

type EmailCampaignListData struct {
	Campaigns []models.EmailCampaign `json:"campaigns"`
	Total     int64                  `json:"total"`
	Limit     int                    `json:"limit"`
	Offset    int                    `json:"offset"`
}

type EmailCampaignSingleResponse struct {
	BaseResponse
	Data EmailCampaignData `json:"data"`
}

type EmailCampaignData struct {
	Campaign models.EmailCampaign `json:"campaign"`
}

type EmailCampaignSendResponse struct {
	BaseResponse
	Data EmailCampaignSendData `json:"data"`
}

type EmailCampaignSendData struct {
	CampaignID     uint `json:"campaign_id"`
	RecipientCount int  `json:"recipient_count"`
	EnqueuedCount  int  `json:"enqueued_count"`
	FailedCount    int  `json:"failed_count"`
}
