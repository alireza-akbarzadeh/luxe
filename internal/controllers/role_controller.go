package controllers

import (
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type RoleController struct {
	service  services.RoleServiceInterface
	validate *validator.Validate
}

func NewRoleController(service services.RoleServiceInterface) *RoleController {
	return &RoleController{
		service:  service,
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
func (ctrl *RoleController) ListRoles(c *gin.Context) {
	roles, err := ctrl.service.ListRoles(c.Request.Context())
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
func (ctrl *RoleController) GetRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "invalid role id")
		return
	}
	role, err := ctrl.service.GetRole(c.Request.Context(), uint(id))
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
func (ctrl *RoleController) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	role, err := ctrl.service.CreateRole(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create role")
		return
	}
	utils.Created(c, "role created", role)
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
func (ctrl *RoleController) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "invalid role id")
		return
	}
	var req dto.UpdateRoleRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	role, err := ctrl.service.UpdateRole(c.Request.Context(), uint(id), &req)
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
func (ctrl *RoleController) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "invalid role id")
		return
	}
	if err := ctrl.service.DeleteRole(c.Request.Context(), uint(id)); err != nil {
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
func (ctrl *RoleController) ListPermissions(c *gin.Context) {
	permissions, err := ctrl.service.ListPermissions(c.Request.Context())
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
func (ctrl *RoleController) SetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "invalid role id")
		return
	}
	var req dto.SetRolePermissionsRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	role, err := ctrl.service.SetRolePermissions(c.Request.Context(), uint(id), req.PermissionIDs)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update role permissions")
		return
	}
	utils.SuccessResponse(c, "role permissions updated", role)
}
