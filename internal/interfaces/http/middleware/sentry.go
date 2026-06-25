package middleware

import (
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// SentryMiddleware captures panics per request (requires observability.Init before routes).
// Repanic is true so gin.Recovery can still return HTTP 500.
func SentryMiddleware() gin.HandlerFunc {
	return sentrygin.New(sentrygin.Options{
		Repanic:         true,
		WaitForDelivery: false,
	})
}

// SentryScope attaches request-scoped tags to the Sentry hub.
// Must run after SentryMiddleware (hub is not available before sentrygin).
func SentryScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		hub := sentrygin.GetHubFromContext(c)
		if hub == nil {
			c.Next()
			return
		}

		if id, ok := c.Get(string(constants.RequestIDKey)); ok {
			hub.Scope().SetTag("request_id", fmt.Sprint(id))
		}
		hub.Scope().SetTag("path", c.FullPath())
		if c.FullPath() == "" {
			hub.Scope().SetTag("path", c.Request.URL.Path)
		}

		c.Next()
	}
}

// CaptureException reports an error using the request hub (not the global Sentry client).
func CaptureException(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if hub := sentrygin.GetHubFromContext(c); hub != nil {
		hub.CaptureException(err)
	}
}
