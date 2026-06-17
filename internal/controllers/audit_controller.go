package controllers

import (
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuditController struct {
	auditService services.AuditServiceInterface
	validate     *validator.Validate
}

func NewAuditController(auditService services.AuditServiceInterface) *AuditController {
	return &AuditController{
		auditService: auditService,
		validate:     validator.New(),
	}
}

// List returns paginated audit logs (admin only).
func (ctrl *AuditController) List(c *gin.Context) {
	var filters dto.AuditLogListFilters
	if !utils.BindAndValidateQuery(c, &filters, ctrl.validate) {
		return
	}

	limit := filters.Limit
	if limit == 0 {
		limit = 20
	}

	logs, total, err := ctrl.auditService.List(c.Request.Context(), limit, filters.Offset)
	if err != nil {
		utils.HandleAppError(c, err, "failed to list audit logs")
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
