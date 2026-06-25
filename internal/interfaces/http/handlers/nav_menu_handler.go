package handlers

import (
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	appnavmenu "github.com/alireza-akbarzadeh/luxe/internal/application/navmenu"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type NavMenuHandler struct {
	commands *appnavmenu.Commands
	queries  *appnavmenu.Queries
	validate *validator.Validate
}

func NewNavMenuHandler(commands *appnavmenu.Commands, queries *appnavmenu.Queries) *NavMenuHandler {
	return &NavMenuHandler{
		commands: commands,
		queries:  queries,
		validate: validator.New(),
	}
}

// GetAll godoc
// @Summary Get all navigation menus
// @Tags Navigation
// @Produce json
// @Success 200 {object} utils.Response{data=[]dto.NavItemResponse}
// @Router /nav-menus [get]
func (ctrl *NavMenuHandler) GetAll(c *gin.Context) {
	menu, err := ctrl.queries.GetAll(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch all menu")
		return
	}
	utils.SuccessResponse(c, "menus retrieved", menu)
}

// GetByID godoc
// @Summary Get one navigation menu
// @Tags Navigation
// @Param id path int true "Menu ID"
// @Success 200 {object} utils.Response{data=dto.NavItemResponse}
// @Router /nav-menus/{id} [get]
func (ctrl *NavMenuHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleServiceError(c, err, "invalid id param")
		return
	}
	menu, err := ctrl.queries.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch menu")
		return
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
func (ctrl *NavMenuHandler) Create(c *gin.Context) {
	var req dto.UpsertNavMenuRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.HandleServiceError(c, err, "invalid nav menu")
		return
	}
	created, err := ctrl.commands.Create(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create menu")
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
func (ctrl *NavMenuHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.HandleServiceError(c, err, "invalid id param")
		return
	}
	var req dto.UpsertNavMenuRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	updated, err := ctrl.commands.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update menu")
		return
	}
	utils.SuccessResponse(c, "menu updated", updated)
}

// Reorder godoc
// @Summary Reorder navigation menus
// @Tags Navigation
// @Accept json
// @Produce json
// @Param body body dto.ReorderNavMenusRequest true "Ordered nav item ids"
// @Success 200 {object} utils.Response
// @Router /nav-menus/reorder [put]
func (ctrl *NavMenuHandler) Reorder(c *gin.Context) {
	var req dto.ReorderNavMenusRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	if err := ctrl.commands.Reorder(c.Request.Context(), &req); err != nil {
		utils.HandleServiceError(c, err, "failed to reorder menus")
		return
	}
	utils.SuccessResponse(c, "menus reordered", nil)
}

// Delete godoc
// @Summary Delete navigation menu
// @Tags Navigation
// @Param id path int true "Menu ID"
// @Success 200 {object} utils.Response
// @Router /nav-menus/{id} [delete]
func (ctrl *NavMenuHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)

	if err != nil {
		utils.HandleServiceError(c, err, "invalid id param")
		return
	}
	err = ctrl.commands.Delete(c.Request.Context(), uint(id))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to delete menu")
		return
	}
	utils.SuccessResponse(c, "menu deleted", nil)
}
