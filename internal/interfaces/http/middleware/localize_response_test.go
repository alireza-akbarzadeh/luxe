package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

func TestLocalizeResponseTranslatesDtoJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if err := i18n.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	router := gin.New()
	router.Use(middleware.Locale())
	router.Use(middleware.LocalizeResponse())
	router.POST("/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": constants.ErrInvalidCredentials,
		})
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.Header.Set("Accept-Language", "fa")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	msg, _ := body["message"].(string)
	if msg != "ایمیل یا رمز عبور نامعتبر است" {
		t.Fatalf("message = %q", msg)
	}
}
