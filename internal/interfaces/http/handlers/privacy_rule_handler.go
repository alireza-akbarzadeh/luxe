package handlers

import (
	"net/http"

	"github.com/alireza-akbarzadeh/luxe/internal/application/privacyrule"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// PrivacyRuleHandler handles privacy rule HTTP requests.
type PrivacyRuleHandler struct {
	service  *privacyrule.Service
	validate *validator.Validate
}

// NewPrivacyRuleHandler creates a new PrivacyRuleHandler.
func NewPrivacyRuleHandler(service *privacyrule.Service) *PrivacyRuleHandler {
	return &PrivacyRuleHandler{
		service:  service,
		validate: validator.New(),
	}
}

// ListActivePrivacyRules godoc
// @Summary      List active privacy rules
// @Description  Returns active privacy rules for apps to fetch and parse markdown. Filter by provider, locale, or key.
// @Tags         privacy-rules
// @Produce      json
// @Param        page      query     int     false  "Page number"     default(1)
// @Param        limit     query     int     false  "Items per page"  default(20)
// @Param        provider  query     string  false  "Filter by provider"
// @Param        locale    query     string  false  "Filter by locale"
// @Param        key       query     string  false  "Filter by key"
// @Success      200  {object}  dto.PrivacyRuleListResponse
// @Failure      500  {object}  utils.Response
// @Router       /privacy-rules [get]
func (h *PrivacyRuleHandler) ListActivePrivacyRules(c *gin.Context) {
	var req dto.ListPrivacyRulesRequest
	if !utils.BindAndValidateQuery(c, &req, h.validate) {
		return
	}

	rules, total, err := h.service.ListActivePublic(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list privacy rules")
		return
	}

	c.JSON(http.StatusOK, dto.PrivacyRuleListResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgFetchSuccess,
			Code:    http.StatusOK,
		},
		Data: dto.PrivacyRuleListData{
			Rules: rules,
			Total: total,
			Page:  req.Page,
			Limit: req.Limit,
		},
	})
}

// GetActivePrivacyRuleByKey godoc
// @Summary      Get active privacy rule by key
// @Description  Returns a single active privacy rule by stable key for cross-app parsing of content_markdown.
// @Tags         privacy-rules
// @Produce      json
// @Param        key     path      string  true   "Rule key"
// @Param        locale  query     string  false  "Locale" default(en)
// @Success      200  {object}  utils.Response{data=dto.PrivacyRuleResponse}
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /privacy-rules/key/{key} [get]
func (h *PrivacyRuleHandler) GetActivePrivacyRuleByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		utils.BadRequestResponse(c, "key is required")
		return
	}
	locale := c.DefaultQuery("locale", "en")

	rule, err := h.service.GetActiveByKey(c.Request.Context(), key, locale)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to retrieve privacy rule")
		return
	}
	utils.SuccessResponse(c, "privacy rule retrieved", rule)
}

// ListActivePrivacyRulesByProvider godoc
// @Summary      List active privacy rules by provider
// @Description  Returns active rules for a provider (includes provider=all). Markdown is returned for client-side parsing.
// @Tags         privacy-rules
// @Produce      json
// @Param        provider  path      string  true   "Provider code"
// @Param        locale    query     string  false  "Locale" default(en)
// @Success      200  {object}  utils.Response{data=[]dto.PrivacyRuleResponse}
// @Failure      500  {object}  utils.Response
// @Router       /privacy-rules/provider/{provider} [get]
func (h *PrivacyRuleHandler) ListActivePrivacyRulesByProvider(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		utils.BadRequestResponse(c, "provider is required")
		return
	}
	locale := c.DefaultQuery("locale", "en")

	rules, err := h.service.ListActiveByProvider(c.Request.Context(), provider, locale)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list privacy rules")
		return
	}
	utils.SuccessResponse(c, "privacy rules retrieved", rules)
}

// AdminListPrivacyRules godoc
// @Summary      Admin list privacy rules
// @Description  Paginated list of all privacy rules with optional filters.
// @Tags         privacy-rules
// @Produce      json
// @Param        page      query     int     false  "Page number"     default(1)
// @Param        limit     query     int     false  "Items per page"  default(20)
// @Param        search    query     string  false  "Search name, key, summary"
// @Param        status    query     string  false  "Filter by status"
// @Param        provider  query     string  false  "Filter by provider"
// @Param        locale    query     string  false  "Filter by locale"
// @Success      200  {object}  dto.PrivacyRuleListResponse
// @Failure      500  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/privacy-rules [get]
func (h *PrivacyRuleHandler) AdminListPrivacyRules(c *gin.Context) {
	var req dto.ListPrivacyRulesRequest
	if !utils.BindAndValidateQuery(c, &req, h.validate) {
		return
	}

	rules, total, err := h.service.List(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list privacy rules")
		return
	}

	c.JSON(http.StatusOK, dto.PrivacyRuleListResponse{
		BaseResponse: dto.BaseResponse{
			Success: true,
			Message: constants.MsgFetchSuccess,
			Code:    http.StatusOK,
		},
		Data: dto.PrivacyRuleListData{
			Rules: rules,
			Total: total,
			Page:  req.Page,
			Limit: req.Limit,
		},
	})
}

// AdminGetPrivacyRule godoc
// @Summary      Admin get privacy rule
// @Tags         privacy-rules
// @Produce      json
// @Param        id   path      int  true  "Privacy rule ID"
// @Success      200  {object}  utils.Response{data=dto.PrivacyRuleResponse}
// @Failure      404  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/privacy-rules/{id} [get]
func (h *PrivacyRuleHandler) AdminGetPrivacyRule(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	rule, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to retrieve privacy rule")
		return
	}
	utils.SuccessResponse(c, "privacy rule retrieved", rule)
}

// AdminCreatePrivacyRule godoc
// @Summary      Create privacy rule
// @Description  Creates a privacy rule with markdown content, provider scope, and optional status (default draft).
// @Tags         privacy-rules
// @Accept       json
// @Produce      json
// @Param        request body dto.CreatePrivacyRuleRequest true "Privacy rule payload"
// @Success      201  {object}  utils.Response{data=dto.PrivacyRuleResponse}
// @Failure      400  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/privacy-rules [post]
func (h *PrivacyRuleHandler) AdminCreatePrivacyRule(c *gin.Context) {
	var req dto.CreatePrivacyRuleRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}

	rule, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create privacy rule")
		return
	}
	utils.CreatedResponse(c, "privacy rule created successfully", rule)
}

// AdminUpdatePrivacyRule godoc
// @Summary      Update privacy rule
// @Tags         privacy-rules
// @Accept       json
// @Produce      json
// @Param        id      path      int                          true  "Privacy rule ID"
// @Param        request body      dto.UpdatePrivacyRuleRequest true  "Update payload"
// @Success      200  {object}  utils.Response{data=dto.PrivacyRuleResponse}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/privacy-rules/{id} [put]
func (h *PrivacyRuleHandler) AdminUpdatePrivacyRule(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdatePrivacyRuleRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}

	rule, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update privacy rule")
		return
	}
	utils.SuccessResponse(c, "privacy rule updated successfully", rule)
}

// AdminDeletePrivacyRule godoc
// @Summary      Delete privacy rule
// @Tags         privacy-rules
// @Produce      json
// @Param        id   path      int  true  "Privacy rule ID"
// @Success      200  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Security     BearerAuth
// @Router       /admin/privacy-rules/{id} [delete]
func (h *PrivacyRuleHandler) AdminDeletePrivacyRule(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		utils.HandleServiceError(c, err, "failed to delete privacy rule")
		return
	}
	utils.SuccessResponse(c, "privacy rule deleted successfully", nil)
}
