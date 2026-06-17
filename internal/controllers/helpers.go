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
func RespondServiceError(c *gin.Context, err error, logMessage string) {
	utils.HandleServiceError(c, err, logMessage)
}
