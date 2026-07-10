package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

func setupTeamRoutes(admin *gin.RouterGroup, ctrl *handlers.Container) {
	admin.GET("/teams", ctrl.Team.ListTeams)
	admin.POST("/teams", ctrl.Team.CreateTeam)
	admin.GET("/teams/:id", ctrl.Team.GetTeam)
	admin.PUT("/teams/:id", ctrl.Team.UpdateTeam)
	admin.DELETE("/teams/:id", ctrl.Team.DeleteTeam)
	admin.POST("/teams/:id/members", ctrl.Team.AddTeamMember)
	admin.DELETE("/teams/:id/members/:userId", ctrl.Team.RemoveTeamMember)
}
