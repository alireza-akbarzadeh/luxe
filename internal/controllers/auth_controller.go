package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthController struct {
	authService services.AuthServiceInterface
	validate    *validator.Validate
}

func NewAuthController(authService services.AuthServiceInterface) *AuthController {
	return &AuthController{
		authService: authService,
		validate:    validator.New(),
	}
}

// Register handles user registration.
// @Summary      Register a new user
// @Description  Create a new account and returns a pair of JWT tokens
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Registration data"
// @Success      201 {object} dto.RegisterResponse
// @Failure      400 {object} dto.MessageResponse
// @Failure      409 {object} dto.MessageResponse
// @Router       /auth/register [post]
func (ctrl *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	accessToken, refreshToken, user, err := ctrl.authService.Register(req, sessionMetaFromContext(c))
	if err != nil {
		utils.HandleAppError(c, err, constants.MsgRegistrationFailed)
		return
	}

	resp := dto.RegisterResponse{
		Success: true,
		Message: constants.MsgRegistrationSuccess,
		Data: dto.RegisterResponseData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User: dto.UserResponse{
				ID:        user.ID,
				Email:     user.Email,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Role:      user.Role,
				Phone:     user.Phone,
			},
		},
	}
	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate and return access & refresh tokens
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login credentials"
// @Success      200 {object} dto.LoginResponse
// @Failure      400 {object} dto.MessageResponse
// @Failure      401 {object} dto.MessageResponse
// @Router       /auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	accessToken, refreshToken, user, err := ctrl.authService.Login(req, sessionMetaFromContext(c))
	if err != nil {
		utils.HandleAppError(c, err, constants.MsgLoginFailed)
		return
	}

	resp := dto.LoginResponse{
		Success: true,
		Message: constants.MsgLoginSuccess,
		Data: dto.LoginResponseData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User: dto.UserResponse{
				ID:        user.ID,
				Email:     user.Email,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Role:      user.Role,
				Phone:     user.Phone,
			},
		},
	}
	c.JSON(http.StatusOK, resp)
}

// RefreshRequest represents the body for token refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Refresh generates a new access token using a valid refresh token.
// @Summary      Refresh access token
// @Description  Obtain a new access token using a valid refresh token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body RefreshRequest true "Refresh token"
// @Success      200 {object} dto.RefreshResponse
// @Failure      400 {object} dto.MessageResponse
// @Failure      401 {object} dto.MessageResponse
// @Router       /auth/refresh [post]
func (ctrl *AuthController) Refresh(c *gin.Context) {
	var req RefreshRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	newAccessToken, newRefreshToken, err := ctrl.authService.RefreshTokens(req.RefreshToken, sessionMetaFromContext(c))
	if err != nil {
		utils.HandleAppError(c, err, "failed to refresh tokens")
		return
	}
	refreshExpiry := config.AppConfig.JWT.RefreshTokenExpiry
	isProduction := strings.EqualFold(config.AppConfig.Server.Mode, "release")

	c.SetCookie(
		"refresh_token",
		newRefreshToken,
		int(refreshExpiry.Seconds()), // maxAge in seconds
		"/",                          // path
		"",                           // domain (current domain)
		isProduction,                 // secure (HTTPS only in production)
		true,                         // httpOnly
	)

	c.Header("Set-Cookie", fmt.Sprintf("%s; SameSite=Lax", c.Writer.Header().Get("Set-Cookie")))

	resp := dto.RefreshResponse{
		Success: true,
		Message: constants.MsgRefreshSuccess,
		Data: dto.RefreshResponseData{
			AccessToken:  newAccessToken,
			RefreshToken: newRefreshToken,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// Logout revokes the user's refresh token(s).
// @Summary      Logout user
// @Description  Invalidate the refresh token (optional specific token)
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object false "Optional {refresh_token}"
// @Success      200 {object} dto.MessageResponse
// @Failure      401 {object} dto.MessageResponse
// @Router       /auth/logout [post]
func (ctrl *AuthController) Logout(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req services.LogoutRequest
	_ = c.ShouldBindJSON(&req)

	if err := ctrl.authService.Logout(userID, req); err != nil {
		utils.HandleServiceError(c, err, "logout failed")
		return
	}

	resp := dto.MessageResponse{
		Success: true,
		Message: constants.MsgLogoutSuccess,
	}
	c.JSON(http.StatusOK, resp)
}

// ChangePassword handles password update for authenticated user.
// @Summary      Change password
// @Description  Change current user's password.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.ChangePasswordRequest true "Password change request"
// @Success      200 {object} dto.MessageResponse
// @Failure      400 {object} dto.MessageResponse
// @Failure      401 {object} dto.MessageResponse
// @Failure      500 {object} dto.MessageResponse
// @Router       /auth/change-password [post]
func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	var req dto.ChangePasswordRequest
	// Fixed: pass &req (was req)
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	err := ctrl.authService.ChangePassword(userID, req)
	if err != nil {
		utils.HandleAppError(c, err, "failed to change password")
		return
	}
	resp := dto.MessageResponse{
		Success: true,
		Message: "password changed successfully",
	}
	c.JSON(http.StatusOK, resp)
}

// ResetPassword handles password reset using token.
// @Summary      Reset password
// @Description  Resets password using a valid token received by email.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordRequest true "Reset token and new password"
// @Success      200 {object} dto.MessageResponse
// @Failure      400 {object} dto.MessageResponse
// @Router       /auth/reset-password [post]
func (ctrl *AuthController) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	err := ctrl.authService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		utils.HandleAppError(c, err, "reset password failed")
		return
	}
	resp := dto.MessageResponse{
		Success: true,
		Message: "password reset successfully",
	}
	c.JSON(http.StatusOK, resp)
}

// ForgotPassword sends reset link.
// @Summary      Forgot password
// @Description  Sends a password reset email
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.ForgotPasswordRequest true "Email address"
// @Success      200 {object} dto.MessageResponse
// @Failure      400 {object} dto.MessageResponse
// @Router       /auth/forgot-password [post]
func (ctrl *AuthController) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	_ = ctrl.authService.ForgotPassword(req.Email)
	// Always return success to avoid email enumeration
	resp := dto.MessageResponse{
		Success: true,
		Message: "if the email exists, you will receive a reset link",
	}
	c.JSON(http.StatusOK, resp)
}

// VerifyEmail verifies user's email address with a token.
// @Summary      Verify email
// @Description  Verifies email address using a token sent via email.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        token query string true "Verification token"
// @Success      200 {object} dto.MessageResponse
// @Failure      400 {object} dto.MessageResponse
// @Router       /auth/verify-email [get]
func (ctrl *AuthController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "token is required")
		return
	}
	err := ctrl.authService.VerifyEmail(token)
	if err != nil {
		utils.HandleAppError(c, err, "email verification failed")
		return
	}
	resp := dto.MessageResponse{
		Success: true,
		Message: "email verified successfully",
	}
	c.JSON(http.StatusOK, resp)
}

// SendVerificationEmail sends a verification email to the authenticated user.
// @Summary      Send verification email
// @Description  Sends an email verification link to the authenticated user's email.
// @Tags         Authentication
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200 {object} dto.MessageResponse
// @Failure      401 {object} dto.MessageResponse
// @Failure      500 {object} dto.MessageResponse
// @Router       /auth/send-verification [post]
func (ctrl *AuthController) SendVerificationEmail(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "user not authenticated")
		return
	}
	err := ctrl.authService.SendVerificationEmail(userID)
	if err != nil {
		utils.HandleAppError(c, err, "failed to send verification email")
		return
	}
	resp := dto.MessageResponse{
		Success: true,
		Message: "verification email sent",
	}
	c.JSON(http.StatusOK, resp)
}

func sessionMetaFromContext(c *gin.Context) services.SessionMeta {
	ip := c.ClientIP()
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			ip = strings.TrimSpace(parts[0])
		}
	}
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		ip = realIP
	}

	return services.SessionMeta{
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: ip,
	}
}

// ListSessions returns active refresh-token sessions for the authenticated user.
func (ctrl *AuthController) ListSessions(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	currentRefreshToken, _ := c.Cookie("refresh_token")
	sessions, err := ctrl.authService.ListSessions(userID, currentRefreshToken)
	if err != nil {
		utils.HandleAppError(c, err, "failed to list sessions")
		return
	}

	resp := dto.SessionsResponse{
		Success: true,
		Message: "sessions retrieved",
		Data: dto.SessionsResponseData{
			Sessions: sessions,
		},
	}
	c.JSON(http.StatusOK, resp)
}

// RevokeSession revokes a single session by ID.
func (ctrl *AuthController) RevokeSession(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	var sessionID uint
	if _, err := fmt.Sscan(c.Param("id"), &sessionID); err != nil || sessionID == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid session id")
		return
	}

	if err := ctrl.authService.RevokeSession(userID, sessionID); err != nil {
		utils.HandleAppError(c, err, "failed to revoke session")
		return
	}

	resp := dto.MessageResponse{
		Success: true,
		Message: "session revoked",
	}
	c.JSON(http.StatusOK, resp)
}

// RevokeOtherSessions revokes all sessions except the current refresh token.
func (ctrl *AuthController) RevokeOtherSessions(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	var req dto.RevokeSessionsRequest
	_ = c.ShouldBindJSON(&req)

	currentRefreshToken := req.RefreshToken
	if currentRefreshToken == "" {
		currentRefreshToken, _ = c.Cookie("refresh_token")
	}

	if err := ctrl.authService.RevokeOtherSessions(userID, currentRefreshToken); err != nil {
		utils.HandleAppError(c, err, "failed to revoke other sessions")
		return
	}

	resp := dto.MessageResponse{
		Success: true,
		Message: "other sessions revoked",
	}
	c.JSON(http.StatusOK, resp)
}
