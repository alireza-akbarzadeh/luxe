package handlers

import (
	appstore "github.com/alireza-akbarzadeh/luxe/internal/application/store"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

// authorizeVendorStore verifies the authenticated user may access the store in the URL.
func authorizeVendorStore(c *gin.Context, storeQueries *appstore.Queries) (storeID, userID uint, role string, ok bool) {
	userID, ok = middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return 0, 0, "", false
	}
	storeID, ok = parseUintParam(c, "id")
	if !ok {
		return 0, 0, "", false
	}
	role, _ = middleware.GetUserRole(c)
	if _, err := storeQueries.GetVendorStore(c.Request.Context(), storeID, userID, role); err != nil {
		RespondServiceError(c, err, "store not found")
		return 0, 0, "", false
	}
	return storeID, userID, role, true
}
