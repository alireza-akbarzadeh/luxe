package handlers

import (
	"net/http"

	"github.com/alireza-akbarzadeh/luxe/internal/shared/health"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	checker *health.Checker
}

func NewHealthHandler(checker *health.Checker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// Check godoc
// @Summary      Health check
// @Description  Readiness-style check (database + Redis when configured). Same as /health/ready.
// @Tags         health
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Failure      503  {object} map[string]interface{}
// @Router       /health [get]
func (hc *HealthHandler) Check(c *gin.Context) {
	hc.respondReady(c)
}

// Live godoc
// @Summary      Liveness probe
// @Description  Returns 200 if the process is alive (no dependency checks).
// @Tags         health
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Router       /health/live [get]
func (hc *HealthHandler) Live(c *gin.Context) {
	status := hc.checker.Live()
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "alive",
		"checks":  status.Checks,
	})
}

// Ready godoc
// @Summary      Readiness probe
// @Description  Returns 200 when database (and Redis if configured) are reachable.
// @Tags         health
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Failure      503  {object} map[string]interface{}
// @Router       /health/ready [get]
func (hc *HealthHandler) Ready(c *gin.Context) {
	hc.respondReady(c)
}

func (hc *HealthHandler) respondReady(c *gin.Context) {
	status := hc.checker.Ready(c.Request.Context())

	response := gin.H{
		"status":  "ok",
		"message": "ready",
		"checks":  status.Checks,
	}
	if !status.OK {
		if !status.Checks["database"] {
			utils.Log.Error("readiness check: database unavailable")
		}
		if c, ok := status.Checks["redis"]; ok && !c {
			utils.Log.Error("readiness check: redis unavailable")
		}
		response["status"] = "degraded"
		response["message"] = "not ready"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}
