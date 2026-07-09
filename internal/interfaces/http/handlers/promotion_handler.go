package handlers

import (
	"net/http"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	apppromotion "github.com/alireza-akbarzadeh/luxe/internal/application/promotion"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// PromotionHandler handles admin merchandising promotion endpoints.
type PromotionHandler struct {
	svc      *apppromotion.Service
	validate *validator.Validate
}

// NewPromotionHandler creates a PromotionHandler.
func NewPromotionHandler(svc *apppromotion.Service) *PromotionHandler {
	return &PromotionHandler{svc: svc, validate: validator.New()}
}

// GetKPIs returns promotion hub KPI counts.
// @Summary      Promotion KPIs
// @Description  Returns active flash deals, published banners, and campaign counts
// @Tags         Promotions
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.PromotionsKPIResponse
// @Router       /admin/promotions/kpis [get]
func (h *PromotionHandler) GetKPIs(c *gin.Context) {
	data, err := h.svc.GetKPIs(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load promotion KPIs")
		return
	}
	c.JSON(http.StatusOK, dto.PromotionsKPIResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data:         *data,
	})
}

// ListFlashDeals lists flash deals for admin.
// @Summary      List flash deals
// @Tags         Promotions
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query int    false "Items per page"
// @Param        offset  query int    false "Offset"
// @Param        status  query string false "Filter by status"
// @Param        search  query string false "Search title or product id"
// @Success      200 {object} dto.FlashDealListResponse
// @Router       /admin/flash-deals [get]
func (h *PromotionHandler) ListFlashDeals(c *gin.Context) {
	var filters dto.AdminFlashDealListFilters
	if !utils.BindAndValidateQuery(c, &filters, h.validate) {
		return
	}
	deals, total, err := h.svc.ListFlashDeals(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list flash deals")
		return
	}
	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	c.JSON(http.StatusOK, dto.FlashDealListResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data: dto.FlashDealListData{
			Deals: deals, Total: total, Limit: limit, Offset: filters.Offset,
		},
	})
}

// GetFlashDeal returns a flash deal by ID.
// @Summary      Get flash deal
// @Tags         Promotions
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Flash deal ID"
// @Success      200 {object} dto.FlashDealSingleResponse
// @Router       /admin/flash-deals/{id} [get]
func (h *PromotionHandler) GetFlashDeal(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	deal, err := h.svc.GetFlashDeal(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get flash deal")
		return
	}
	c.JSON(http.StatusOK, dto.FlashDealSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data:         dto.FlashDealData{Deal: *deal},
	})
}

// CreateFlashDeal creates a flash deal.
// @Summary      Create flash deal
// @Tags         Promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateFlashDealRequest true "Flash deal payload"
// @Success      201 {object} dto.FlashDealSingleResponse
// @Router       /admin/flash-deals [post]
func (h *PromotionHandler) CreateFlashDeal(c *gin.Context) {
	var req dto.CreateFlashDealRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	deal, err := h.svc.CreateFlashDeal(c.Request.Context(), req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create flash deal")
		return
	}
	c.JSON(http.StatusCreated, dto.FlashDealSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgCreateSuccess, Code: http.StatusCreated},
		Data:         dto.FlashDealData{Deal: *deal},
	})
}

// UpdateFlashDeal updates a flash deal.
// @Summary      Update flash deal
// @Tags         Promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Flash deal ID"
// @Param        request body dto.UpdateFlashDealRequest true "Flash deal update"
// @Success      200 {object} dto.FlashDealSingleResponse
// @Router       /admin/flash-deals/{id} [put]
func (h *PromotionHandler) UpdateFlashDeal(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateFlashDealRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	deal, err := h.svc.UpdateFlashDeal(c.Request.Context(), id, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update flash deal")
		return
	}
	c.JSON(http.StatusOK, dto.FlashDealSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgUpdateSuccess, Code: http.StatusOK},
		Data:         dto.FlashDealData{Deal: *deal},
	})
}

// DeleteFlashDeal deletes a flash deal.
// @Summary      Delete flash deal
// @Tags         Promotions
// @Security     BearerAuth
// @Param        id path int true "Flash deal ID"
// @Success      200 {object} dto.EmptyResponse
// @Router       /admin/flash-deals/{id} [delete]
func (h *PromotionHandler) DeleteFlashDeal(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteFlashDeal(c.Request.Context(), id); err != nil {
		utils.HandleServiceError(c, err, "failed to delete flash deal")
		return
	}
	c.JSON(http.StatusOK, dto.EmptyResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgDeleteSuccess, Code: http.StatusOK},
	})
}

// ListHomepageSections lists featured banners.
// @Summary      List homepage sections
// @Tags         Promotions
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query int    false "Items per page"
// @Param        offset  query int    false "Offset"
// @Param        status  query string false "Filter by status"
// @Param        search  query string false "Search title or key"
// @Success      200 {object} dto.HomepageSectionListResponse
// @Router       /admin/homepage-sections [get]
func (h *PromotionHandler) ListHomepageSections(c *gin.Context) {
	var filters dto.AdminHomepageSectionListFilters
	if !utils.BindAndValidateQuery(c, &filters, h.validate) {
		return
	}
	sections, total, err := h.svc.ListHomepageSections(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list homepage sections")
		return
	}
	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	c.JSON(http.StatusOK, dto.HomepageSectionListResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data: dto.HomepageSectionListData{
			Sections: sections, Total: total, Limit: limit, Offset: filters.Offset,
		},
	})
}

// GetHomepageSection returns a homepage section by ID.
// @Summary      Get homepage section
// @Tags         Promotions
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Section ID"
// @Success      200 {object} dto.HomepageSectionSingleResponse
// @Router       /admin/homepage-sections/{id} [get]
func (h *PromotionHandler) GetHomepageSection(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	section, err := h.svc.GetHomepageSection(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get homepage section")
		return
	}
	c.JSON(http.StatusOK, dto.HomepageSectionSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data:         dto.HomepageSectionData{Section: *section},
	})
}

// CreateHomepageSection creates a featured banner section.
// @Summary      Create homepage section
// @Tags         Promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateHomepageSectionRequest true "Section payload"
// @Success      201 {object} dto.HomepageSectionSingleResponse
// @Router       /admin/homepage-sections [post]
func (h *PromotionHandler) CreateHomepageSection(c *gin.Context) {
	var req dto.CreateHomepageSectionRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	section, err := h.svc.CreateHomepageSection(c.Request.Context(), req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create homepage section")
		return
	}
	c.JSON(http.StatusCreated, dto.HomepageSectionSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgCreateSuccess, Code: http.StatusCreated},
		Data:         dto.HomepageSectionData{Section: *section},
	})
}

// UpdateHomepageSection updates a homepage section.
// @Summary      Update homepage section
// @Tags         Promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Section ID"
// @Param        request body dto.UpdateHomepageSectionRequest true "Section update"
// @Success      200 {object} dto.HomepageSectionSingleResponse
// @Router       /admin/homepage-sections/{id} [put]
func (h *PromotionHandler) UpdateHomepageSection(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateHomepageSectionRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	section, err := h.svc.UpdateHomepageSection(c.Request.Context(), id, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update homepage section")
		return
	}
	c.JSON(http.StatusOK, dto.HomepageSectionSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgUpdateSuccess, Code: http.StatusOK},
		Data:         dto.HomepageSectionData{Section: *section},
	})
}

// DeleteHomepageSection deletes a homepage section.
// @Summary      Delete homepage section
// @Tags         Promotions
// @Security     BearerAuth
// @Param        id path int true "Section ID"
// @Success      200 {object} dto.EmptyResponse
// @Router       /admin/homepage-sections/{id} [delete]
func (h *PromotionHandler) DeleteHomepageSection(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteHomepageSection(c.Request.Context(), id); err != nil {
		utils.HandleServiceError(c, err, "failed to delete homepage section")
		return
	}
	c.JSON(http.StatusOK, dto.EmptyResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgDeleteSuccess, Code: http.StatusOK},
	})
}

// ListCampaigns lists merchandising campaigns.
// @Summary      List campaigns
// @Tags         Promotions
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query int    false "Items per page"
// @Param        offset  query int    false "Offset"
// @Param        status  query string false "Filter by status"
// @Param        search  query string false "Search name or slug"
// @Success      200 {object} dto.CampaignListResponse
// @Router       /admin/campaigns [get]
func (h *PromotionHandler) ListCampaigns(c *gin.Context) {
	var filters dto.AdminCampaignListFilters
	if !utils.BindAndValidateQuery(c, &filters, h.validate) {
		return
	}
	campaigns, total, err := h.svc.ListCampaigns(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list campaigns")
		return
	}
	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	c.JSON(http.StatusOK, dto.CampaignListResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data: dto.CampaignListData{
			Campaigns: campaigns, Total: total, Limit: limit, Offset: filters.Offset,
		},
	})
}

// GetCampaign returns a campaign by ID.
// @Summary      Get campaign
// @Tags         Promotions
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Campaign ID"
// @Success      200 {object} dto.CampaignSingleResponse
// @Router       /admin/campaigns/{id} [get]
func (h *PromotionHandler) GetCampaign(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	campaign, err := h.svc.GetCampaign(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get campaign")
		return
	}
	c.JSON(http.StatusOK, dto.CampaignSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data:         dto.CampaignData{Campaign: *campaign},
	})
}

// CreateCampaign creates a merchandising campaign.
// @Summary      Create campaign
// @Tags         Promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateCampaignRequest true "Campaign payload"
// @Success      201 {object} dto.CampaignSingleResponse
// @Router       /admin/campaigns [post]
func (h *PromotionHandler) CreateCampaign(c *gin.Context) {
	var req dto.CreateCampaignRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	campaign, err := h.svc.CreateCampaign(c.Request.Context(), req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create campaign")
		return
	}
	c.JSON(http.StatusCreated, dto.CampaignSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgCreateSuccess, Code: http.StatusCreated},
		Data:         dto.CampaignData{Campaign: *campaign},
	})
}

// UpdateCampaign updates a campaign.
// @Summary      Update campaign
// @Tags         Promotions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Campaign ID"
// @Param        request body dto.UpdateCampaignRequest true "Campaign update"
// @Success      200 {object} dto.CampaignSingleResponse
// @Router       /admin/campaigns/{id} [put]
func (h *PromotionHandler) UpdateCampaign(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateCampaignRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	campaign, err := h.svc.UpdateCampaign(c.Request.Context(), id, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update campaign")
		return
	}
	c.JSON(http.StatusOK, dto.CampaignSingleResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgUpdateSuccess, Code: http.StatusOK},
		Data:         dto.CampaignData{Campaign: *campaign},
	})
}

// DeleteCampaign deletes a campaign.
// @Summary      Delete campaign
// @Tags         Promotions
// @Security     BearerAuth
// @Param        id path int true "Campaign ID"
// @Success      200 {object} dto.EmptyResponse
// @Router       /admin/campaigns/{id} [delete]
func (h *PromotionHandler) DeleteCampaign(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteCampaign(c.Request.Context(), id); err != nil {
		utils.HandleServiceError(c, err, "failed to delete campaign")
		return
	}
	c.JSON(http.StatusOK, dto.EmptyResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgDeleteSuccess, Code: http.StatusOK},
	})
}
