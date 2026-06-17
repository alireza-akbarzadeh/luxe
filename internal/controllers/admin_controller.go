package controllers

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
)

type AdminController struct {
	adminService services.AdminServiceInterface
}

func NewAdminController(svc services.AdminServiceInterface) *AdminController {
	return &AdminController{adminService: svc}
}

// GetStats returns platform-wide statistics (admin only).
// @Summary      Platform stats (admin)
// @Description  Returns counts of users, orders, products, revenue, and wallet balances.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.AdminStatsResponse}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/stats [get]
func (ctrl *AdminController) GetStats(c *gin.Context) {
	stats, err := ctrl.adminService.GetStats(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get stats")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, stats)
}
