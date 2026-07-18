package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	appauth "github.com/alireza-akbarzadeh/luxe/internal/application/auth"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type authServicer interface {
	Register(ctx context.Context, req dto.RegisterRequest, meta appauth.SessionMeta) (string, string, *models.User, error)
	Login(ctx context.Context, req dto.LoginRequest, meta appauth.SessionMeta) (string, string, *models.User, error)
	RequestLoginOTP(ctx context.Context, identifier string) (dto.RequestLoginOTPData, error)
	VerifyLoginOTP(ctx context.Context, identifier, code string, meta appauth.SessionMeta) (string, string, *models.User, error)
	RefreshTokens(ctx context.Context, rawRefreshToken string, meta appauth.SessionMeta) (newAccessToken, newRawRefreshToken string, err error)
	Logout(ctx context.Context, userID uint, req appauth.LogoutRequest) error
	ChangePassword(ctx context.Context, userID uint, req dto.ChangePasswordRequest) error
	ResetPassword(ctx context.Context, token string, newPassword string) error
	ForgotPassword(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, token string) error
	SendVerificationEmail(ctx context.Context, userID uint) error
	ListSessions(ctx context.Context, userID uint, currentRefreshToken string) ([]dto.SessionResponse, error)
	RevokeSession(ctx context.Context, userID uint, sessionID uint) error
	RevokeOtherSessions(ctx context.Context, userID uint, currentRefreshToken string) error
}

type AuthHandler struct {
	authService authServicer
	validate    *validator.Validate
}

func NewAuthHandler(authService authServicer) *AuthHandler {
	return &AuthHandler{
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
func (ctrl *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	accessToken, refreshToken, user, err := ctrl.authService.Register(c.Request.Context(), req, sessionMetaFromContext(c))
	if err != nil {
		utils.HandleServiceError(c, err, constants.MsgRegistrationFailed)
		return
	}

	resp := dto.RegisterResponse{
		Success: true,
		Message: constants.MsgRegistrationSuccess,
		Data: dto.RegisterResponseData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User: dto.ToUserResponse(user),
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
func (ctrl *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	accessToken, refreshToken, user, err := ctrl.authService.Login(c.Request.Context(), req, sessionMetaFromContext(c))
	if err != nil {
		utils.HandleServiceError(c, err, constants.MsgLoginFailed)
		return
	}

	resp := dto.LoginResponse{
		Success: true,
		Message: constants.MsgLoginSuccess,
		Data: dto.LoginResponseData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User: dto.ToUserResponse(user),
		},
	}
	c.JSON(http.StatusOK, resp)
}

// RequestLoginOTP sends a one-time sign-in code to the account email.
// @Summary      Request login OTP
// @Description  Sends a 6-digit code by email. Identifier may be email or E.164 phone (phone delivers to the account email).
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.RequestLoginOTPRequest true "Email or phone"
// @Success      200 {object} utils.Response{data=dto.RequestLoginOTPData}
// @Failure      400 {object} utils.Response
// @Router       /auth/login/otp/request [post]
func (ctrl *AuthHandler) RequestLoginOTP(c *gin.Context) {
	var req dto.RequestLoginOTPRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	data, err := ctrl.authService.RequestLoginOTP(c.Request.Context(), req.Identifier)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to send login code")
		return
	}
	utils.SuccessResponse(c, "If an account exists, a sign-in code was sent", data)
}

// VerifyLoginOTP completes passwordless login with the emailed code.
// @Summary      Verify login OTP
// @Description  Exchanges a valid 6-digit code for access and refresh tokens
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.VerifyLoginOTPRequest true "Identifier + code"
// @Success      200 {object} dto.LoginResponse
// @Failure      400 {object} dto.MessageResponse
// @Failure      401 {object} dto.MessageResponse
// @Router       /auth/login/otp/verify [post]
func (ctrl *AuthHandler) VerifyLoginOTP(c *gin.Context) {
	var req dto.VerifyLoginOTPRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	accessToken, refreshToken, user, err := ctrl.authService.VerifyLoginOTP(
		c.Request.Context(),
		req.Identifier,
		req.Code,
		sessionMetaFromContext(c),
	)
	if err != nil {
		utils.HandleServiceError(c, err, constants.MsgLoginFailed)
		return
	}

	resp := dto.LoginResponse{
		Success: true,
		Message: constants.MsgLoginSuccess,
		Data: dto.LoginResponseData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User:         dto.ToUserResponse(user),
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
func (ctrl *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	newAccessToken, newRefreshToken, err := ctrl.authService.RefreshTokens(c.Request.Context(), req.RefreshToken, sessionMetaFromContext(c))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to refresh tokens")
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
func (ctrl *AuthHandler) Logout(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req appauth.LogoutRequest
	_ = c.ShouldBindJSON(&req)

	if err := ctrl.authService.Logout(c.Request.Context(), userID, req); err != nil {
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
func (ctrl *AuthHandler) ChangePassword(c *gin.Context) {
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
	err := ctrl.authService.ChangePassword(c.Request.Context(), userID, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to change password")
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
func (ctrl *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	err := ctrl.authService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		utils.HandleServiceError(c, err, "reset password failed")
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
func (ctrl *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	_ = ctrl.authService.ForgotPassword(c.Request.Context(), req.Email)
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
func (ctrl *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "token is required")
		return
	}
	err := ctrl.authService.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		utils.HandleServiceError(c, err, "email verification failed")
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
func (ctrl *AuthHandler) SendVerificationEmail(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "user not authenticated")
		return
	}
	err := ctrl.authService.SendVerificationEmail(c.Request.Context(), userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to send verification email")
		return
	}
	resp := dto.MessageResponse{
		Success: true,
		Message: "verification email sent",
	}
	c.JSON(http.StatusOK, resp)
}

func sessionMetaFromContext(c *gin.Context) appauth.SessionMeta {
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

	return appauth.SessionMeta{
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: ip,
	}
}

// ListSessions returns active refresh-token sessions for the authenticated user.
func (ctrl *AuthHandler) ListSessions(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	currentRefreshToken, _ := c.Cookie("refresh_token")
	sessions, err := ctrl.authService.ListSessions(c.Request.Context(), userID, currentRefreshToken)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list sessions")
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
func (ctrl *AuthHandler) RevokeSession(c *gin.Context) {
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

	if err := ctrl.authService.RevokeSession(c.Request.Context(), userID, sessionID); err != nil {
		utils.HandleServiceError(c, err, "failed to revoke session")
		return
	}

	resp := dto.MessageResponse{
		Success: true,
		Message: "session revoked",
	}
	c.JSON(http.StatusOK, resp)
}

// RevokeOtherSessions revokes all sessions except the current refresh token.
func (ctrl *AuthHandler) RevokeOtherSessions(c *gin.Context) {
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

	if err := ctrl.authService.RevokeOtherSessions(c.Request.Context(), userID, currentRefreshToken); err != nil {
		utils.HandleServiceError(c, err, "failed to revoke other sessions")
		return
	}

	resp := dto.MessageResponse{
		Success: true,
		Message: "other sessions revoked",
	}
	c.JSON(http.StatusOK, resp)
}
