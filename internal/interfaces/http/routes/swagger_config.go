package routes

import (
	"net/url"
	"os"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/docs"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
)

// configureSwaggerInfo sets Swagger host/schemes from deployment env so "Try it out"
// on Render hits the live API instead of localhost:8080 baked in at swag init time.
func configureSwaggerInfo(cfg *config.Config) {
	host := strings.TrimSpace(os.Getenv("SWAGGER_HOST"))
	schemes := []string{"http", "https"}

	if host == "" {
		if renderURL := strings.TrimSpace(os.Getenv("RENDER_EXTERNAL_URL")); renderURL != "" {
			if u, err := url.Parse(renderURL); err == nil && u.Host != "" {
				host = u.Host
				if u.Scheme == "https" {
					schemes = []string{"https", "http"}
				}
			}
		}
	}

	if host == "" {
		port := strings.TrimSpace(cfg.Server.Port)
		if port == "" {
			port = "8080"
		}
		host = "localhost:" + port
	}

	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.Schemes = schemes
	docs.SwaggerInfo.BasePath = "/api/v1"
}
