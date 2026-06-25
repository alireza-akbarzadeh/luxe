package handlers

import (
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type MenuHandler struct {
	menuService services.UserMenuServicesInterface
	validate    *validator.Validate
}

func NewMenuHandler(menuService services.UserMenuServicesInterface) *MenuHandler {
	return &MenuHandler{
		menuService: menuService,
		validate:    validator.New(),
	}
}

// GetAllGroups returns all menu groups.
// @Summary      Get all menu groups
// @Description  Retrieves all menu groups ordered by display_order
// @Tags         Admin Menu Groups
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=[]models.MenuGroup}
// @Failure      500 {object} utils.Response
// @Router       /admin/menu/groups [get]
func (ctrl *MenuHandler) GetAllGroups(c *gin.Context) {
	groups, err := ctrl.menuService.GetAllGroups()
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch menu groups")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, groups)
}

// GetGroupByID returns a single menu group by ID.
// @Summary      Get group by ID
// @Description  Returns a single menu group by its ID
// @Tags         Admin Menu Groups
// @Produce      json
// @Param        id   path      int  true  "Group ID"
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=models.MenuGroup}
// @Failure      400 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/menu/groups/{id} [get]
func (ctrl *MenuHandler) GetGroupByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid group id")
		return
	}
	group, err := ctrl.menuService.GetGroupByID(uint(id))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch group")
		return
	}
	if group == nil {
		utils.NotFoundResponse(c, "group not found")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, group)
}

// CreateGroup creates a new menu group.
// @Summary      Create a new menu group
// @Description  Creates a menu group (e.g., "Overview", "Users & Access")
// @Tags         Admin Menu Groups
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateMenuGroupRequest true "Group data"
// @Security     BearerAuth
// @Success      201 {object} utils.Response{data=models.MenuGroup}
// @Failure      400 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/menu/groups [post]
func (ctrl *MenuHandler) CreateGroup(c *gin.Context) {
	var req dto.CreateMenuGroupRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	group, err := ctrl.menuService.CreateGroup(&req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create group")
		return
	}
	utils.CreatedResponse(c, constants.MsgCreateSuccess, group)
}

// UpdateGroup updates an existing menu group.
// @Summary      Update an existing menu group
// @Description  Updates group name or display order
// @Tags         Admin Menu Groups
// @Accept       json
// @Produce      json
// @Param        id       path      int                        true "Group ID"
// @Param        request  body      dto.UpdateMenuGroupRequest true "Updated group data"
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=models.MenuGroup}
// @Failure      400 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/menu/groups/{id} [put]
func (ctrl *MenuHandler) UpdateGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid group id")
		return
	}
	var req dto.UpdateMenuGroupRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	group, err := ctrl.menuService.UpdateGroup(uint(id), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update group")
		return
	}
	utils.SuccessResponse(c, constants.MsgUpdateSuccess, group)
}

// DeleteGroup deletes a menu group.
// @Summary      Delete a menu group
// @Description  Deletes a group and all its menu items (cascade)
// @Tags         Admin Menu Groups
// @Produce      json
// @Param        id   path      int  true "Group ID"
// @Security     BearerAuth
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/menu/groups/{id} [delete]
func (ctrl *MenuHandler) DeleteGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid group id")
		return
	}
	if err := ctrl.menuService.DeleteGroup(uint(id)); err != nil {
		utils.HandleServiceError(c, err, "failed to delete group")
		return
	}
	utils.SuccessResponse(c, constants.MsgDeleteSuccess, nil)
}

// GetAllItems returns all menu items (flat or nested).
// @Summary      Get all menu items
// @Description  Returns flat or nested menu items (use ?flat=true for flat list)
// @Tags         Admin Menu Items
// @Produce      json
// @Param        flat  query   bool  false  "Return flat list" default(false)
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.MenuListResponse}
// @Failure      500 {object} utils.Response
// @Router       /admin/menu/items [get]
func (ctrl *MenuHandler) GetAllItems(c *gin.Context) {
	flat, _ := strconv.ParseBool(c.DefaultQuery("flat", "false"))
	items, err := ctrl.menuService.GetAllItems(flat)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch menu items")
		return
	}
	resp := dto.MenuListResponse{
		Items: items,
		BaseResponse: dto.BaseResponse{
			Success: true,
			Code:    200,
			Message: constants.MsgFetchSuccess,
		},
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, resp)
}

// GetItemByID returns a single menu item.
// @Summary      Get menu item by ID
// @Description  Returns a single menu item by its ID
// @Tags         Admin Menu Items
// @Produce      json
// @Param        id   path      int  true "Item ID"
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=models.MenuItem}
// @Failure      400 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/menu/items/{id} [get]
func (ctrl *MenuHandler) GetItemByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid item id")
		return
	}
	item, err := ctrl.menuService.GetItemByID(uint(id))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch item")
		return
	}
	if item == nil {
		utils.NotFoundResponse(c, "item not found")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, item)
}

// CreateItem creates a new menu item.
// @Summary      Create a new menu item
// @Description  Adds a new menu item (can be top-level or child of another item)
// @Tags         Admin Menu Items
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateMenuItemRequest true "Menu item data"
// @Security     BearerAuth
// @Success      201 {object} utils.Response{data=models.MenuItem}
// @Failure      400 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/menu/items [post]
func (ctrl *MenuHandler) CreateItem(c *gin.Context) {
	var req dto.CreateMenuItemRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	item, err := ctrl.menuService.CreateItem(&req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create item")
		return
	}
	utils.CreatedResponse(c, constants.MsgCreateSuccess, item)
}

// UpdateItem updates a menu item.
// @Summary      Update an existing menu item
// @Description  Updates menu item details including group, parent, label, href, etc.
// @Tags         Admin Menu Items
// @Accept       json
// @Produce      json
// @Param        id       path      int                       true "Item ID"
// @Param        request  body      dto.UpdateMenuItemRequest true "Updated item data"
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=models.MenuItem}
// @Failure      400 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/menu/items/{id} [put]
func (ctrl *MenuHandler) UpdateItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid item id")
		return
	}
	var req dto.UpdateMenuItemRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	item, err := ctrl.menuService.UpdateItem(uint(id), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update item")
		return
	}
	utils.SuccessResponse(c, constants.MsgUpdateSuccess, item)
}

// DeleteItem deletes a menu item.
// @Summary      Delete a menu item
// @Description  Deletes a menu item and all its children (cascade)
// @Tags         Admin Menu Items
// @Produce      json
// @Param        id   path      int  true  "Menu item ID"
// @Security     BearerAuth
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/menu/items/{id} [delete]
func (ctrl *MenuHandler) DeleteItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid item id")
		return
	}
	if err := ctrl.menuService.DeleteItem(uint(id)); err != nil {
		utils.HandleServiceError(c, err, "failed to delete item")
		return
	}
	utils.SuccessResponse(c, constants.MsgDeleteSuccess, nil)
}

// GetUserMenu returns the sidebar menu for the current user.
// @Summary      Get user sidebar menu
// @Description  Returns the sidebar menu filtered by user's role and optional search term
// @Tags         User Menu
// @Produce      json
// @Param        search  query   string  false  "Search by label or href"
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=[]dto.SidebarGroup}
// @Failure      500 {object} utils.Response
// @Router       /user/menu [get]
func (ctrl *MenuHandler) GetUserMenu(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists {
		userRole = "guest"
	}
	search := c.Query("search")
	menu, err := ctrl.menuService.GetUserMenu(c.Request.Context(), userRole.(string), search)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch menu")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, menu)
}

// GetUserMenuStructure returns groups + nested items filtered by user role.
// @Summary      Get user menu structure (grouped + nested)
// @Description  Returns all menu groups and their items, nested and ordered, filtered by user's permissions.
// @Tags         User Menu
// @Produce      json
// @Param        search  query   string  false  "Search by label or href"
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=[]dto.MenuGroupResponse}
// @Failure      500 {object} utils.Response
// @Router       /user/menu/structure [get]
func (ctrl *MenuHandler) GetUserMenuStructure(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists {
		userRole = "guest"
	}
	search := c.Query("search")
	structure, err := ctrl.menuService.GetUserMenuStructure(c.Request.Context(), userRole.(string), search)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch user menu structure")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, structure)
}
