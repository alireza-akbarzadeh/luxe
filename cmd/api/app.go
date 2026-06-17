package main

import "github.com/gin-gonic/gin"

// setupGin creates the Gin engine with recovery middleware.
// HTTP access logs are handled by middleware.AccessLog (structured JSON).
func setupGin() *gin.Engine {
	engine := gin.New()
	engine.RedirectTrailingSlash = false
	engine.Use(gin.Recovery())
	return engine
}
