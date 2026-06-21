package middleware

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
	"github.com/gin-gonic/gin"
)

// Locale resolves Accept-Language and stores the locale on the Gin and request contexts.
func Locale() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := i18n.ParseAcceptLanguage(c.GetHeader("Accept-Language"))
		c.Set(string(constants.LocaleContextKey), locale)
		c.Request = c.Request.WithContext(i18n.WithLocale(c.Request.Context(), locale))
		c.Next()
	}
}
