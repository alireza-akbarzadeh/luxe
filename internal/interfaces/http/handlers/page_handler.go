package handlers

import (
	"github.com/alireza-akbarzadeh/luxe/views"
	"github.com/gin-gonic/gin"
)

type PageHandler struct{}

func NewPageHandler() *PageHandler {
	return &PageHandler{}
}

// LandingPage serves the main HTML page.
func (pc *PageHandler) LandingPage(c *gin.Context) {
	if err := views.RenderTemplate(c.Writer, "index.html", nil); err != nil {
		c.String(500, err.Error())
		return
	}
}
