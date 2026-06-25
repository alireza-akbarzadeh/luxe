package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequireAdmin_AllowsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("user_role", constants.RoleAdmin)

	RequireAdmin()(c)
	assert.False(t, c.IsAborted())
}

func TestRequireAdmin_BlocksUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("user_role", constants.RoleUser)

	RequireAdmin()(c)
	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestIsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_role", constants.RoleAdmin)
	assert.True(t, IsAdmin(c))

	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Set("user_role", constants.RoleUser)
	assert.False(t, IsAdmin(c2))
}
