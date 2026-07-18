package handlers

import (
	appaudit "github.com/alireza-akbarzadeh/luxe/internal/application/audit"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuditHandler struct {
	queries  *appaudit.Queries
	validate *validator.Validate
}

func NewAuditHandler(queries *appaudit.Queries) *AuditHandler {
	return &AuditHandler{
		queries:  queries,
		validate: validator.New(),
	}
}

// List returns paginated audit logs (admin only).
// @Summary List audit logs
// @Tags Admin Audit
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Page size"
// @Param offset query int false "Offset"
// @Param search query string false "Search path, resource, email"
// @Param action query string false "HTTP action"
// @Param resource query string false "Resource path filter"
// @Param user_id query int false "Actor user id"
// @Param date_from query string false "From date (YYYY-MM-DD or RFC3339)"
// @Param date_to query string false "To date (YYYY-MM-DD or RFC3339)"
// @Success 200 {object} utils.Response
// @Router /admin/audit-logs [get]
func (ctrl *AuditHandler) List(c *gin.Context) {
	var filters dto.AuditLogListFilters
	if !utils.BindAndValidateQuery(c, &filters, ctrl.validate) {
		return
	}

	limit := filters.Limit
	if limit == 0 {
		limit = 20
	}

	logs, total, err := ctrl.queries.List(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list audit logs")
		return
	}

	items := make([]dto.AuditLogResponse, len(logs))
	for i, log := range logs {
		items[i] = dto.AuditLogResponse{
			ID:         log.ID,
			UserID:     log.UserID,
			Action:     log.Action,
			Resource:   log.Resource,
			ResourceID: log.ResourceID,
			Path:       log.Path,
			IPAddress:  log.IPAddress,
			RequestID:  log.RequestID,
			CreatedAt:  log.CreatedAt,
			UserEmail:  log.User.Email,
		}
	}

	utils.SuccessResponse(c, "audit logs retrieved", gin.H{
		"logs":   items,
		"total":  total,
		"limit":  limit,
		"offset": filters.Offset,
	})
}

// Summary returns aggregate audit metrics.
// @Summary Audit log summary
// @Tags Admin Audit
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=dto.AuditLogSummaryResponse}
// @Router /admin/audit-logs/summary [get]
func (ctrl *AuditHandler) Summary(c *gin.Context) {
	summary, err := ctrl.queries.Summary(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load audit summary")
		return
	}
	utils.SuccessResponse(c, "audit summary retrieved", summary)
}
