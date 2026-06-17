package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// uint test

type MockAuthService struct {
	mock.Mock
	services.AuthServiceInterface
}

func (m *MockAuthService) Login(ctx context.Context, req dto.LoginRequest, meta services.SessionMeta) (string, string, *models.User, error) {
	args := m.Called(ctx, req, meta)
	return args.String(0), args.String(1), args.Get(2).(*models.User), args.Error(3)
}

func (m *MockAuthService) Register(ctx context.Context, req dto.RegisterRequest, meta services.SessionMeta) (string, string, *models.User, error) {
	args := m.Called(ctx, req, meta)
	return args.String(0), args.String(1), args.Get(2).(*models.User), args.Error(3)
}

func (m *MockAuthService) RefreshTokens(ctx context.Context, refreshToken string, meta services.SessionMeta) (string, string, error) {
	args := m.Called(ctx, refreshToken, meta)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) Logout(ctx context.Context, userID uint, req services.LogoutRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *MockAuthService) ChangePassword(ctx context.Context, userID uint, req dto.ChangePasswordRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *MockAuthService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	args := m.Called(ctx, token, newPassword)
	return args.Error(0)
}

func (m *MockAuthService) ForgotPassword(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func TestAuthController_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuthService := new(MockAuthService)
	ctrl := NewAuthController(mockAuthService)

	t.Run("success login return 200", func(t *testing.T) {
		expectedUser := &models.User{
			ID:        1,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Role:      "user",
			Phone:     "1234567890",
		}
		mockAuthService.On("Login", mock.Anything, dto.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}, mock.Anything).Return("access_token", "refresh_token", expectedUser, nil).Once()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"email":"test@example.com","password":"password123"}`
		c.Request = httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")
		ctrl.Login(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp dto.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "access_token", resp.Data.AccessToken)
		assert.Equal(t, "refresh_token", resp.Data.RefreshToken)
		assert.Equal(t, expectedUser.ID, resp.Data.User.ID)
		mockAuthService.AssertExpectations(t)
	})
	t.Run("invalid credential return 401", func(t *testing.T) {
		mockAuthService.On("Login", mock.Anything, dto.LoginRequest{
			Email:    "wrong@example.com",
			Password: "wrong",
		}, mock.Anything).Return("", "", (*models.User)(nil), utils.ErrUnauthorized("invalid email or password")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"email":"wrong@example.com","password":"wrong"}`
		c.Request = httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")

		ctrl.Login(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp dto.MessageResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "invalid")
		mockAuthService.AssertExpectations(t)
	})
	t.Run("validation error returns 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"email":"not-an-email","password":"123"}`
		c.Request = httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")

		ctrl.Login(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp dto.MessageResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "validation")
	})
}
