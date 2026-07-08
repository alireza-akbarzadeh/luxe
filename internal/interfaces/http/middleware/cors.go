package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func parseAllowOrigins() []string {
	seen := make(map[string]struct{})
	add := func(origin string) {
		origin = strings.TrimSpace(strings.TrimRight(origin, "/"))
		if origin != "" {
			seen[origin] = struct{}{}
		}
	}

	raw := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS"))
	if raw == "" {
		for _, origin := range defaultAllowOrigins() {
			add(origin)
		}
	} else {
		for _, part := range strings.Split(raw, ",") {
			add(part)
		}
	}

	// Always allow configured storefront origin when set separately.
	add(os.Getenv("FRONTEND_URL"))

	origins := make([]string, 0, len(seen))
	for origin := range seen {
		origins = append(origins, origin)
	}

	return origins
}

func defaultAllowOrigins() []string {
	return []string{
		"http://localhost:3000",
		"https://luxe-3pvz.onrender.com",
		"http://127.0.0.1:3000",
		"http://localhost:4000",
		"http://127.0.0.1:4000",
		// Expo dev (Metro web + legacy web port)
		"http://localhost:8081",
		"http://127.0.0.1:8081",
		"http://localhost:19006",
		"http://127.0.0.1:19006",
	}
}

// AllowedOrigins returns CORS-allowed origins (shared with WebSocket origin checks).
func AllowedOrigins() []string {
	return parseAllowOrigins()
}

// CORS returns a Gin middleware that enables Cross-Origin Resource Sharing.
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: parseAllowOrigins(),
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"Cache-Control",
		},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
