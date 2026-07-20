package handlers

import (
	"net/http"
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

func computeDiscountPercent(original, discounted float64) *int {
	if discounted <= 0 || discounted >= original {
		return nil
	}
	percent := int(((original - discounted) / original) * 100)
	return &percent
}

// RespondServiceError maps service-layer errors (utils.AppError) to HTTP responses.
func RespondServiceError(c *gin.Context, err error, logMessage string) {
	utils.HandleServiceError(c, err, logMessage)
}

// paginationParams extracts and clamps limit/offset from query params.
// defaultLimit is used when the caller does not send a value; max is always capped at constants.MaxLimit.
func paginationParams(c *gin.Context, defaultLimit int) (limit, offset int) {
	if defaultLimit <= 0 {
		defaultLimit = constants.DefaultLimit
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit < constants.MinLimit {
		limit = defaultLimit
	}
	if limit > constants.MaxLimit {
		limit = constants.MaxLimit
	}

	offset, err = strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}
	return limit, offset
}

// pageLimitParams extracts and clamps page/limit from query params for
// admin listing endpoints that paginate by page number instead of offset.
func pageLimitParams(c *gin.Context, defaultLimit int) (page, limit int) {
	if defaultLimit <= 0 {
		defaultLimit = constants.DefaultLimit
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err = strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit < constants.MinLimit {
		limit = defaultLimit
	}
	if limit > constants.MaxLimit {
		limit = constants.MaxLimit
	}
	return page, limit
}

// parseUintParam reads a named URL parameter as uint. Returns (0, false) on failure
// and writes a 400 response so the caller can just `return`.
func parseUintParam(c *gin.Context, param string) (uint, bool) {
	v, err := strconv.ParseUint(c.Param(param), 10, 64)
	if err != nil || v == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid "+param)
		return 0, false
	}
	return uint(v), true
}
