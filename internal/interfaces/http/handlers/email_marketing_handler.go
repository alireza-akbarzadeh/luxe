package handlers

import (
	"net/http"

	appemailmarketing "github.com/alireza-akbarzadeh/luxe/internal/application/emailmarketing"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// EmailMarketingHandler handles newsletter and email campaign endpoints.
type EmailMarketingHandler struct {
	svc      *appemailmarketing.Service
	validate *validator.Validate
}

// NewEmailMarketingHandler creates an EmailMarketingHandler.
func NewEmailMarketingHandler(svc *appemailmarketing.Service) *EmailMarketingHandler {
	return &EmailMarketingHandler{svc: svc, validate: validator.New()}
}

// GetKPIs returns email marketing hub metrics.
// @Summary      Email marketing KPIs
// @Description  Returns subscriber, template, and campaign analytics
// @Tags         Email Marketing
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.EmailMarketingKPIResponse
// @Router       /admin/email-marketing/kpis [get]
func (h *EmailMarketingHandler) GetKPIs(c *gin.Context) {
	data, err := h.svc.GetKPIs(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load email marketing KPIs")
		return
	}
	c.JSON(http.StatusOK, dto.EmailMarketingKPIResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data:         *data,
	})
}

// Subscribe handles public newsletter signup.
// @Summary      Subscribe to newsletter
// @Tags         Newsletters
// @Accept       json
// @Produce      json
// @Param        request body dto.SubscribeNewsletterRequest true "Subscribe payload"
// @Success      201 {object} dto.SubscriberSingleResponse
// @Router       /newsletters/subscribe [post]
func (h *EmailMarketingHandler) Subscribe(c *gin.Context) {
	var req dto.SubscribeNewsletterRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	var userID *uint
	if uid, ok := middleware.GetUserID(c); ok && uid > 0 {
		userID = &uid
	}
	subscriber, err := h.svc.Subscribe(c.Request.Context(), req, userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to subscribe")
		return
	}
	c.JSON(http.StatusCreated, dto.SubscriberSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: "Subscribed successfully", Code: http.StatusCreated},
		Data:         dto.SubscriberData{Subscriber: *subscriber},
	})
}

// Unsubscribe handles public newsletter opt-out.
// @Summary      Unsubscribe from newsletter
// @Tags         Newsletters
// @Accept       json
// @Produce      json
// @Param        request body dto.UnsubscribeNewsletterRequest true "Unsubscribe token"
// @Success      200 {object} dto.BaseResponse
// @Router       /newsletters/unsubscribe [post]
func (h *EmailMarketingHandler) Unsubscribe(c *gin.Context) {
	var req dto.UnsubscribeNewsletterRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	if err := h.svc.Unsubscribe(c.Request.Context(), req.Token); err != nil {
		utils.HandleServiceError(c, err, "failed to unsubscribe")
		return
	}
	c.JSON(http.StatusOK, dto.BaseResponse{Success: true, Message: "Unsubscribed successfully", Code: http.StatusOK})
}

// ListSubscribers lists newsletter subscribers for admin.
// @Summary      List newsletter subscribers
// @Tags         Email Marketing
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query int    false "Items per page"
// @Param        offset  query int    false "Offset"
// @Param        status  query string false "Filter by status"
// @Param        source  query string false "Filter by source"
// @Param        search  query string false "Search email"
// @Success      200 {object} dto.SubscriberListResponse
// @Router       /admin/newsletter-subscribers [get]
func (h *EmailMarketingHandler) ListSubscribers(c *gin.Context) {
	var filters dto.AdminSubscriberListFilters
	if !utils.BindAndValidateQuery(c, &filters, h.validate) {
		return
	}
	rows, total, err := h.svc.ListSubscribers(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list subscribers")
		return
	}
	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	c.JSON(http.StatusOK, dto.SubscriberListResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data: dto.SubscriberListData{
			Subscribers: rows, Total: total, Limit: limit, Offset: filters.Offset,
		},
	})
}

// ExportSubscribers exports subscribers as CSV.
// @Summary      Export newsletter subscribers
// @Tags         Email Marketing
// @Produce      text/csv
// @Security     BearerAuth
// @Param        status query string false "Filter by status"
// @Param        source query string false "Filter by source"
// @Param        search query string false "Search email"
// @Success      200 {string} string "CSV file"
// @Router       /admin/newsletter-subscribers/export [get]
func (h *EmailMarketingHandler) ExportSubscribers(c *gin.Context) {
	var filters dto.AdminSubscriberListFilters
	_ = c.ShouldBindQuery(&filters)
	csvData, err := h.svc.ExportSubscribersCSV(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to export subscribers")
		return
	}
	c.Header("Content-Disposition", "attachment; filename=newsletter-subscribers.csv")
	c.Data(http.StatusOK, "text/csv", csvData)
}

// DeleteSubscriber removes a subscriber.
// @Summary      Delete newsletter subscriber
// @Tags         Email Marketing
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Subscriber ID"
// @Success      200 {object} dto.BaseResponse
// @Router       /admin/newsletter-subscribers/{id} [delete]
func (h *EmailMarketingHandler) DeleteSubscriber(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteSubscriber(c.Request.Context(), id); err != nil {
		utils.HandleServiceError(c, err, "failed to delete subscriber")
		return
	}
	c.JSON(http.StatusOK, dto.BaseResponse{Success: true, Message: constants.MsgDeleteSuccess, Code: http.StatusOK})
}

// ListTemplates lists email templates.
// @Summary      List email templates
// @Tags         Email Marketing
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query int    false "Items per page"
// @Param        offset  query int    false "Offset"
// @Param        status  query string false "Filter by status"
// @Param        search  query string false "Search"
// @Success      200 {object} dto.EmailTemplateListResponse
// @Router       /admin/email-templates [get]
func (h *EmailMarketingHandler) ListTemplates(c *gin.Context) {
	var filters dto.AdminEmailTemplateListFilters
	if !utils.BindAndValidateQuery(c, &filters, h.validate) {
		return
	}
	rows, total, err := h.svc.ListTemplates(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list templates")
		return
	}
	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	c.JSON(http.StatusOK, dto.EmailTemplateListResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data: dto.EmailTemplateListData{
			Templates: rows, Total: total, Limit: limit, Offset: filters.Offset,
		},
	})
}

// GetTemplate returns an email template.
// @Summary      Get email template
// @Tags         Email Marketing
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Template ID"
// @Success      200 {object} dto.EmailTemplateSingleResponse
// @Router       /admin/email-templates/{id} [get]
func (h *EmailMarketingHandler) GetTemplate(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	row, err := h.svc.GetTemplate(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get template")
		return
	}
	c.JSON(http.StatusOK, dto.EmailTemplateSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data:         dto.EmailTemplateData{Template: *row},
	})
}

// CreateTemplate creates an email template.
// @Summary      Create email template
// @Tags         Email Marketing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateEmailTemplateRequest true "Template payload"
// @Success      201 {object} dto.EmailTemplateSingleResponse
// @Router       /admin/email-templates [post]
func (h *EmailMarketingHandler) CreateTemplate(c *gin.Context) {
	var req dto.CreateEmailTemplateRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	row, err := h.svc.CreateTemplate(c.Request.Context(), req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create template")
		return
	}
	c.JSON(http.StatusCreated, dto.EmailTemplateSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgCreateSuccess, Code: http.StatusCreated},
		Data:         dto.EmailTemplateData{Template: *row},
	})
}

// UpdateTemplate updates an email template.
// @Summary      Update email template
// @Tags         Email Marketing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Template ID"
// @Param        request body dto.UpdateEmailTemplateRequest true "Template payload"
// @Success      200 {object} dto.EmailTemplateSingleResponse
// @Router       /admin/email-templates/{id} [put]
func (h *EmailMarketingHandler) UpdateTemplate(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateEmailTemplateRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	row, err := h.svc.UpdateTemplate(c.Request.Context(), id, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update template")
		return
	}
	c.JSON(http.StatusOK, dto.EmailTemplateSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgUpdateSuccess, Code: http.StatusOK},
		Data:         dto.EmailTemplateData{Template: *row},
	})
}

// DeleteTemplate deletes an email template.
// @Summary      Delete email template
// @Tags         Email Marketing
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Template ID"
// @Success      200 {object} dto.BaseResponse
// @Router       /admin/email-templates/{id} [delete]
func (h *EmailMarketingHandler) DeleteTemplate(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteTemplate(c.Request.Context(), id); err != nil {
		utils.HandleServiceError(c, err, "failed to delete template")
		return
	}
	c.JSON(http.StatusOK, dto.BaseResponse{Success: true, Message: constants.MsgDeleteSuccess, Code: http.StatusOK})
}

// ListCampaigns lists email campaigns.
// @Summary      List email campaigns
// @Tags         Email Marketing
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query int    false "Items per page"
// @Param        offset  query int    false "Offset"
// @Param        status  query string false "Filter by status"
// @Param        search  query string false "Search"
// @Success      200 {object} dto.EmailCampaignListResponse
// @Router       /admin/email-campaigns [get]
func (h *EmailMarketingHandler) ListCampaigns(c *gin.Context) {
	var filters dto.AdminEmailCampaignListFilters
	if !utils.BindAndValidateQuery(c, &filters, h.validate) {
		return
	}
	rows, total, err := h.svc.ListCampaigns(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list campaigns")
		return
	}
	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	c.JSON(http.StatusOK, dto.EmailCampaignListResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data: dto.EmailCampaignListData{
			Campaigns: rows, Total: total, Limit: limit, Offset: filters.Offset,
		},
	})
}

// GetCampaign returns an email campaign.
// @Summary      Get email campaign
// @Tags         Email Marketing
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Campaign ID"
// @Success      200 {object} dto.EmailCampaignSingleResponse
// @Router       /admin/email-campaigns/{id} [get]
func (h *EmailMarketingHandler) GetCampaign(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	row, err := h.svc.GetCampaign(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get campaign")
		return
	}
	c.JSON(http.StatusOK, dto.EmailCampaignSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data:         dto.EmailCampaignData{Campaign: *row},
	})
}

// CreateCampaign creates an email campaign.
// @Summary      Create email campaign
// @Tags         Email Marketing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateEmailCampaignRequest true "Campaign payload"
// @Success      201 {object} dto.EmailCampaignSingleResponse
// @Router       /admin/email-campaigns [post]
func (h *EmailMarketingHandler) CreateCampaign(c *gin.Context) {
	var req dto.CreateEmailCampaignRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	row, err := h.svc.CreateCampaign(c.Request.Context(), req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create campaign")
		return
	}
	c.JSON(http.StatusCreated, dto.EmailCampaignSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgCreateSuccess, Code: http.StatusCreated},
		Data:         dto.EmailCampaignData{Campaign: *row},
	})
}

// UpdateCampaign updates an email campaign.
// @Summary      Update email campaign
// @Tags         Email Marketing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Campaign ID"
// @Param        request body dto.UpdateEmailCampaignRequest true "Campaign payload"
// @Success      200 {object} dto.EmailCampaignSingleResponse
// @Router       /admin/email-campaigns/{id} [put]
func (h *EmailMarketingHandler) UpdateCampaign(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateEmailCampaignRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	row, err := h.svc.UpdateCampaign(c.Request.Context(), id, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update campaign")
		return
	}
	c.JSON(http.StatusOK, dto.EmailCampaignSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgUpdateSuccess, Code: http.StatusOK},
		Data:         dto.EmailCampaignData{Campaign: *row},
	})
}

// DeleteCampaign deletes an email campaign.
// @Summary      Delete email campaign
// @Tags         Email Marketing
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Campaign ID"
// @Success      200 {object} dto.BaseResponse
// @Router       /admin/email-campaigns/{id} [delete]
func (h *EmailMarketingHandler) DeleteCampaign(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteCampaign(c.Request.Context(), id); err != nil {
		utils.HandleServiceError(c, err, "failed to delete campaign")
		return
	}
	c.JSON(http.StatusOK, dto.BaseResponse{Success: true, Message: constants.MsgDeleteSuccess, Code: http.StatusOK})
}

// ScheduleCampaign schedules an email campaign.
// @Summary      Schedule email campaign
// @Tags         Email Marketing
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Campaign ID"
// @Param        request body dto.ScheduleEmailCampaignRequest true "Schedule payload"
// @Success      200 {object} dto.EmailCampaignSingleResponse
// @Router       /admin/email-campaigns/{id}/schedule [post]
func (h *EmailMarketingHandler) ScheduleCampaign(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.ScheduleEmailCampaignRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	row, err := h.svc.ScheduleCampaign(c.Request.Context(), id, req.ScheduledAt)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to schedule campaign")
		return
	}
	c.JSON(http.StatusOK, dto.EmailCampaignSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgUpdateSuccess, Code: http.StatusOK},
		Data:         dto.EmailCampaignData{Campaign: *row},
	})
}

// SendCampaign sends an email campaign immediately.
// @Summary      Send email campaign
// @Tags         Email Marketing
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Campaign ID"
// @Success      200 {object} dto.EmailCampaignSendResponse
// @Router       /admin/email-campaigns/{id}/send [post]
func (h *EmailMarketingHandler) SendCampaign(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	data, err := h.svc.SendCampaign(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to send campaign")
		return
	}
	c.JSON(http.StatusOK, dto.EmailCampaignSendResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: "Campaign queued for delivery", Code: http.StatusOK},
		Data:         *data,
	})
}
