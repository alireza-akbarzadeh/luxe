package handlers

import (
	"strconv"

	appteam "github.com/alireza-akbarzadeh/luxe/internal/application/team"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type TeamHandler struct {
	commands *appteam.Commands
	queries  *appteam.Queries
	validate *validator.Validate
}

func NewTeamHandler(commands *appteam.Commands, queries *appteam.Queries) *TeamHandler {
	return &TeamHandler{
		commands: commands,
		queries:  queries,
		validate: validator.New(),
	}
}

// ListTeams godoc
// @Summary List teams
// @Tags Admin Teams
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]dto.TeamResponse}
// @Router /admin/teams [get]
func (ctrl *TeamHandler) ListTeams(c *gin.Context) {
	teams, err := ctrl.queries.ListTeams(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list teams")
		return
	}
	utils.SuccessResponse(c, "teams retrieved", teams)
}

// GetTeam godoc
// @Summary Get team with members
// @Tags Admin Teams
// @Produce json
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Success 200 {object} utils.Response{data=dto.TeamResponse}
// @Router /admin/teams/{id} [get]
func (ctrl *TeamHandler) GetTeam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleServiceError(c, err, "invalid id param")
		return
	}
	team, err := ctrl.queries.GetTeam(c.Request.Context(), uint(id))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get team")
		return
	}
	utils.SuccessResponse(c, "team retrieved", team)
}

// CreateTeam godoc
// @Summary Create team
// @Tags Admin Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.CreateTeamRequest true "Team data"
// @Success 201 {object} utils.Response{data=dto.TeamResponse}
// @Router /admin/teams [post]
func (ctrl *TeamHandler) CreateTeam(c *gin.Context) {
	var req dto.CreateTeamRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	team, err := ctrl.commands.CreateTeam(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create team")
		return
	}
	utils.CreatedResponse(c, "team created", team)
}

// UpdateTeam godoc
// @Summary Update team
// @Tags Admin Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Param body body dto.UpdateTeamRequest true "Team data"
// @Success 200 {object} utils.Response{data=dto.TeamResponse}
// @Router /admin/teams/{id} [put]
func (ctrl *TeamHandler) UpdateTeam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleServiceError(c, err, "invalid id param")
		return
	}
	var req dto.UpdateTeamRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	team, err := ctrl.commands.UpdateTeam(c.Request.Context(), uint(id), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update team")
		return
	}
	utils.SuccessResponse(c, "team updated", team)
}

// DeleteTeam godoc
// @Summary Delete team
// @Tags Admin Teams
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Success 200 {object} utils.Response
// @Router /admin/teams/{id} [delete]
func (ctrl *TeamHandler) DeleteTeam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleServiceError(c, err, "invalid id param")
		return
	}
	if err := ctrl.commands.DeleteTeam(c.Request.Context(), uint(id)); err != nil {
		utils.HandleServiceError(c, err, "failed to delete team")
		return
	}
	utils.SuccessResponse(c, "team deleted", nil)
}

// AddTeamMember godoc
// @Summary Add team member
// @Tags Admin Teams
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Param body body dto.AddTeamMemberRequest true "Member data"
// @Success 200 {object} utils.Response{data=dto.TeamResponse}
// @Router /admin/teams/{id}/members [post]
func (ctrl *TeamHandler) AddTeamMember(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleServiceError(c, err, "invalid id param")
		return
	}
	var req dto.AddTeamMemberRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	team, err := ctrl.commands.AddMember(c.Request.Context(), uint(id), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to add team member")
		return
	}
	utils.SuccessResponse(c, "team member added", team)
}

// RemoveTeamMember godoc
// @Summary Remove team member
// @Tags Admin Teams
// @Security BearerAuth
// @Param id path int true "Team ID"
// @Param userId path int true "User ID"
// @Success 200 {object} utils.Response{data=dto.TeamResponse}
// @Router /admin/teams/{id}/members/{userId} [delete]
func (ctrl *TeamHandler) RemoveTeamMember(c *gin.Context) {
	teamID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleServiceError(c, err, "invalid team id param")
		return
	}
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		utils.HandleServiceError(c, err, "invalid user id param")
		return
	}
	team, err := ctrl.commands.RemoveMember(c.Request.Context(), uint(teamID), uint(userID))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to remove team member")
		return
	}
	utils.SuccessResponse(c, "team member removed", team)
}
