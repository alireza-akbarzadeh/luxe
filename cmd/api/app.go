package main

import "github.com/gin-gonic/gin"

// setupGin creates the Gin engine with default middleware.
func setupGin() *gin.Engine {
	engine := gin.New()
	// Trailing-slash 301 redirects omit CORS headers and break browser cross-origin calls.
	engine.RedirectTrailingSlash = false
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	return engine
}
