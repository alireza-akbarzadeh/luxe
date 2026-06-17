package controllers

import (
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
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
// Use for all service call failures instead of InternalServerErrorResponse when the error may be domain-specific.
func RespondServiceError(c *gin.Context, err error, logMessage string) {
	utils.HandleAppError(c, err, logMessage)
}
