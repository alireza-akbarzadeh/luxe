package handlers

import (
	"errors"

	appsettings "github.com/alireza-akbarzadeh/luxe/internal/application/settings"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SettingHandler struct {
	commands *appsettings.Commands
	queries  *appsettings.Queries
	validate *validator.Validate
}

func NewSettingHandler(commands *appsettings.Commands, queries *appsettings.Queries) *SettingHandler {
	return &SettingHandler{
		commands: commands,
		queries:  queries,
		validate: validator.New(),
	}
}

// GetSetting godoc
// @Summary      Get a setting by key
// @Description  Returns the value (any JSON) for the given setting key.
// @Tags         settings
// @Produce      json
// @Param        key  path      string  true  "Setting key"
// @Success      200  {object}  utils.Response{data=dto.SettingResponse}  "Setting found"
// @Failure      404  {object}  utils.Response  "Setting not found"
// @Router       /settings/{key} [get]
func (ctrl *SettingHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	setting, err := ctrl.queries.Get(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, appsettings.ErrNotFound) {
			utils.NotFoundResponse(c, "setting not found")
			return
		}
		utils.HandleServiceError(c, err, "failed to get setting")
		return
	}
	utils.SuccessResponse(c, "setting retrieved", setting)
}

// ListSettings godoc
// @Summary      List all settings
// @Description  Returns a list of all settings with their keys and values.
// @Tags         settings
// @Produce      json
// @Success      200  {object}  utils.Response{data=[]dto.SettingResponse}  "Settings list"
// @Router       /settings [get]
func (ctrl *SettingHandler) ListSettings(c *gin.Context) {
	settings, err := ctrl.queries.List(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list settings")
		return
	}
	utils.SuccessResponse(c, "settings retrieved", settings)
}

// SetSetting godoc
// @Summary      Create or update a setting
// @Description  Upserts a setting by key. If the key exists, value is updated; otherwise created.
// @Tags         settings
// @Accept       json
// @Produce      json
// @Param        key      path      string                    true  "Setting key"
// @Param        request  body      dto.SetSettingRequest     true  "Setting payload"
// @Success      200      {object}  utils.Response{data=dto.SettingResponse}  "Setting updated"
// @Success      201      {object}  utils.Response{data=dto.SettingResponse}  "Setting created"
// @Failure      400      {object}  utils.Response  "Validation error"
// @Router       /settings/{key} [put]
func (ctrl *SettingHandler) SetSetting(c *gin.Context) {
	key := c.Param("key")
	var req dto.SetSettingRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	set, err := ctrl.commands.Set(c.Request.Context(), key, &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to upsert setting")
		return
	}

	// Since it's an upsert, we could return 200 or 201. 200 is safe.
	utils.SuccessResponse(c, "setting saved", set)
}

// DeleteSetting godoc
// @Summary      Delete a setting
// @Description  Deletes a setting by its key. Admin only.
// @Tags         settings
// @Produce      json
// @Param        key  path      string  true  "Setting key"
// @Success      200  {object}  utils.Response  "Setting deleted"
// @Failure      404  {object}  utils.Response  "Setting not found"
// @Router       /settings/{key} [delete]
func (ctrl *SettingHandler) DeleteSetting(c *gin.Context) {
	key := c.Param("key")
	err := ctrl.commands.Delete(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, appsettings.ErrNotFound) {
			utils.NotFoundResponse(c, "setting not found")
			return
		}
		utils.HandleServiceError(c, err, "failed to delete setting")
		return
	}
	utils.SuccessResponse(c, "setting deleted", nil)
}
