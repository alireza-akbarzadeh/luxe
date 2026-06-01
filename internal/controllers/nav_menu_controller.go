package controllers

import (
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type NavMenuController struct {
	service  services.NavMenuServiceInterface
	validate *validator.Validate
}

func NewNavMenuController(service services.NavMenuServiceInterface) *NavMenuController {
	return &NavMenuController{
		service:  service,
		validate: validator.New(),
	}
}

// GetAll godoc
// @Summary Get all navigation menus
// @Tags Navigation
// @Produce json
// @Success 200 {object} utils.Response{data=[]dto.NavItemResponse}
// @Router /nav-menus [get]
func (ctrl *NavMenuController) GetAll(c *gin.Context) {
	menu, err := ctrl.service.GetAll(c.Request.Context())
	if err != nil {
		utils.HandleAppError(c, err, "failed to fetch all menu")
	}
	utils.SuccessResponse(c, "menus retrieved", menu)
}

// GetByID godoc
// @Summary Get one navigation menu
// @Tags Navigation
// @Param id path int true "Menu ID"
// @Success 200 {object} utils.Response{data=dto.NavItemResponse}
// @Router /nav-menus/{id} [get]
func (ctrl *NavMenuController) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleAppError(c, err, "invalid id param")
	}
	menu, err := ctrl.service.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		utils.HandleAppError(c, err, "failed to fetch menu")
	}
	utils.SuccessResponse(c, "menu retrieved", menu)
}

// Create godoc
// @Summary Create navigation menu
// @Tags Navigation
// @Accept json
// @Produce json
// @Param body body dto.UpsertNavMenuRequest true "Menu data"
// @Success 201 {object} utils.Response{data=dto.NavItemResponse}
// @Router /nav-menus [post]
func (ctrl *NavMenuController) Create(c *gin.Context) {
	var req dto.UpsertNavMenuRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.HandleAppError(c, err, "invalid nav menu")
		return
	}
	created, err := ctrl.service.Create(c.Request.Context(), &req)
	if err != nil {
		utils.HandleAppError(c, err, "failed to create menu")
		return
	}
	utils.SuccessResponse(c, "menu created", created)
}

// Update godoc
// @Summary Update navigation menu
// @Tags Navigation
// @Param id path int true "Menu ID"
// @Accept json
// @Produce json
// @Param body body dto.UpsertNavMenuRequest true "Menu data"
// @Success 200 {object} utils.Response{data=dto.NavItemResponse}
// @Router /nav-menus/{id} [put]
func (ctrl *NavMenuController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleAppError(c, err, "invalid id param")
		return
	}
	var req dto.UpsertNavMenuRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	updated, err := ctrl.service.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		utils.HandleAppError(c, err, "failed to update menu")
		return
	}
	utils.SuccessResponse(c, "menu updated", updated)
}

// Delete godoc
// @Summary Delete navigation menu
// @Tags Navigation
// @Param id path int true "Menu ID"
// @Success 200 {object} utils.Response
// @Router /nav-menus/{id} [delete]
func (ctrl *NavMenuController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		utils.HandleAppError(c, err, "invalid id param")
		return
	}
	err = ctrl.service.Delete(c.Request.Context(), uint(id))
	if err != nil {
		utils.HandleAppError(c, err, "failed to delete menu")
		return
	}
	utils.SuccessResponse(c, "menu deleted", nil)
}
