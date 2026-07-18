package handlers

import (
	"strconv"

	approle "github.com/alireza-akbarzadeh/luxe/internal/application/role"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type RoleHandler struct {
	commands *approle.Commands
	queries  *approle.Queries
	validate *validator.Validate
}

func NewRoleHandler(commands *approle.Commands, queries *approle.Queries) *RoleHandler {
	return &RoleHandler{
		commands: commands,
		queries:  queries,
		validate: validator.New(),
	}
}

// ListRoles godoc
// @Summary List roles
// @Tags Admin Roles
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]dto.RoleResponse}
// @Router /admin/roles [get]
func (ctrl *RoleHandler) ListRoles(c *gin.Context) {
	roles, err := ctrl.queries.ListRoles(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list roles")
		return
	}
	utils.SuccessResponse(c, "roles retrieved", roles)
}

// GetRole godoc
// @Summary Get role with permissions
// @Tags Admin Roles
// @Produce json
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Success 200 {object} utils.Response{data=dto.RoleResponse}
// @Router /admin/roles/{id} [get]
func (ctrl *RoleHandler) GetRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleServiceError(c, err, "invalid id param")
		return
	}
	role, err := ctrl.queries.GetRole(c.Request.Context(), uint(id))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get role")
		return
	}
	utils.SuccessResponse(c, "role retrieved", role)
}

// CreateRole godoc
// @Summary Create role
// @Tags Admin Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.CreateRoleRequest true "Role data"
// @Success 201 {object} utils.Response{data=dto.RoleResponse}
// @Router /admin/roles [post]
func (ctrl *RoleHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	role, err := ctrl.commands.CreateRole(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create role")
		return
	}
	utils.CreatedResponse(c, constants.MsgCreateSuccess, role)
}

// UpdateRole godoc
// @Summary Update role
// @Tags Admin Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Param body body dto.UpdateRoleRequest true "Role data"
// @Success 200 {object} utils.Response{data=dto.RoleResponse}
// @Router /admin/roles/{id} [put]
func (ctrl *RoleHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleServiceError(c, err, "invalid id param")
		return
	}
	var req dto.UpdateRoleRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	role, err := ctrl.commands.UpdateRole(c.Request.Context(), uint(id), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update role")
		return
	}
	utils.SuccessResponse(c, "role updated", role)
}

// DeleteRole godoc
// @Summary Delete role
// @Tags Admin Roles
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Success 200 {object} utils.Response
// @Router /admin/roles/{id} [delete]
func (ctrl *RoleHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleServiceError(c, err, "invalid id param")
		return
	}
	if err := ctrl.commands.DeleteRole(c.Request.Context(), uint(id)); err != nil {
		utils.HandleServiceError(c, err, "failed to delete role")
		return
	}
	utils.SuccessResponse(c, "role deleted", nil)
}

// ListPermissions godoc
// @Summary List permissions
// @Tags Admin Roles
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]dto.PermissionResponse}
// @Router /admin/permissions [get]
func (ctrl *RoleHandler) ListPermissions(c *gin.Context) {
	permissions, err := ctrl.queries.ListPermissions(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list permissions")
		return
	}
	utils.SuccessResponse(c, "permissions retrieved", permissions)
}

// SetRolePermissions godoc
// @Summary Replace permissions assigned to a role
// @Tags Admin Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Role ID"
// @Param body body dto.SetRolePermissionsRequest true "Permission ids"
// @Success 200 {object} utils.Response{data=dto.RoleResponse}
// @Router /admin/roles/{id}/permissions [put]
func (ctrl *RoleHandler) SetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleServiceError(c, err, "invalid id param")
		return
	}
	var req dto.SetRolePermissionsRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	role, err := ctrl.commands.SetRolePermissions(c.Request.Context(), uint(id), req.PermissionIDs)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update role permissions")
		return
	}
	utils.SuccessResponse(c, "role permissions updated", role)
}
