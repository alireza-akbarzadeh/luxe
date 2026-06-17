package controllers

import (
	"net/http"

	"github.com/alireza-akbarzadeh/luxe/internal/database"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthController struct {
	db *gorm.DB
}

func NewHealthController(db *gorm.DB) *HealthController {
	return &HealthController{db: db}
}

// Check godoc
// @Summary      Health check
// @Description  Returns the health status of the API and database
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      503  {object}  map[string]interface{}
// @Router       /health [get]
func (hc *HealthController) Check(c *gin.Context) {
	response := gin.H{
		"status":  "ok",
		"message": "Service is up and running",
	}

	if err := database.Ping(hc.db); err != nil {
		utils.Log.WithError(err).Error("health check: database ping failed")
		response["status"] = "degraded"
		response["db_ok"] = false
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	response["db_ok"] = true
	c.JSON(http.StatusOK, response)
}
