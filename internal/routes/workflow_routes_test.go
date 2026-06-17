package routes

import (
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/controllers"
	"github.com/gin-gonic/gin"
)

// TestSetupWorkflowRoutes_NoPanic ensures the workflow route tree registers without
// Gin radix-tree conflicts (static vs. param segment collisions).
func TestSetupWorkflowRoutes_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	group := engine.Group("/api/v1")

	ctrl := &controllers.Container{
		Workflow: controllers.NewWorkflowController(nil),
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("route registration panicked: %v", r)
		}
	}()

	SetupWorkflowRoutes(group, ctrl)
}
