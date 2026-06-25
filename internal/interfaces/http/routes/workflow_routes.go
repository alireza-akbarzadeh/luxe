package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupWorkflowRoutes registers generic (authenticated) workflow endpoints plus
// admin-only definition CRUD.
func SetupWorkflowRoutes(protected *gin.RouterGroup, ctrl *handlers.Container) {
	wf := protected.Group("/workflows")
	{
		wf.GET("/:key", ctrl.Workflow.GetDefinition)
		wf.POST("/:key/:entityId/transition", ctrl.Workflow.PerformTransition)
		wf.GET("/:key/:entityId/available-transitions", ctrl.Workflow.AvailableTransitions)
		// History uses entityType in the :key slot (same path shape, different handler set below).
		wf.GET("/:key/:entityId/history", ctrl.Workflow.History)
	}

	admin := protected.Group("/admin/workflows")
	admin.Use(middleware.ModuleGuard("settings"))
	{
		admin.GET("", ctrl.Workflow.ListWorkflows)
		admin.POST("", ctrl.Workflow.CreateWorkflow)
		admin.PATCH("/:id", ctrl.Workflow.UpdateWorkflow)
		admin.DELETE("/:id", ctrl.Workflow.DeleteWorkflow)

		// States and transitions are nested under the workflow id to keep Gin's
		// radix tree free of static/param segment conflicts.
		admin.POST("/:id/states", ctrl.Workflow.CreateState)
		admin.PATCH("/:id/states/:stateId", ctrl.Workflow.UpdateState)
		admin.DELETE("/:id/states/:stateId", ctrl.Workflow.DeleteState)

		admin.POST("/:id/transitions", ctrl.Workflow.CreateTransition)
		admin.PATCH("/:id/transitions/:transitionId", ctrl.Workflow.UpdateTransition)
		admin.DELETE("/:id/transitions/:transitionId", ctrl.Workflow.DeleteTransition)
	}
}
