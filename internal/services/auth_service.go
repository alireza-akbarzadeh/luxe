package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/tasks"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

const refreshTokenReuseGracePeriod = 30 * time.Second

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

type SessionMeta struct {
	UserAgent string
	IPAddress string
}

type AuthServiceInterface interface {
	Register(ctx context.Context, req dto.RegisterRequest, meta SessionMeta) (accessToken, refreshToken string, user *models.User, err error)
	Login(ctx context.Context, req dto.LoginRequest, meta SessionMeta) (accessToken, refreshToken string, user *models.User, err error)
	RefreshTokens(ctx context.Context, refreshToken string, meta SessionMeta) (newAccessToken, newRefreshToken string, err error)
	Logout(ctx context.Context, userID uint, req LogoutRequest) error
	ListSessions(ctx context.Context, userID uint, currentRefreshToken string) ([]dto.SessionResponse, error)
	RevokeSession(ctx context.Context, userID, sessionID uint) error
	RevokeOtherSessions(ctx context.Context, userID uint, currentRefreshToken string) error
	VerifyEmail(ctx context.Context, token string) error
	ChangePassword(ctx context.Context, userID uint, req dto.ChangePasswordRequest) error
	ResetPassword(ctx context.Context, token string, newPassword string) error
	ForgotPassword(ctx context.Context, email string) error
	SendVerificationEmail(ctx context.Context, userID uint) error
}
type AuthService struct {
	db       *gorm.DB
	cfg      *config.Config
	jobQueue tasks.JobQueue
}

func NewAuthServices(db *gorm.DB, cfg *config.Config, jobQueue tasks.JobQueue) *AuthService {
	return &AuthService{db: db, cfg: cfg, jobQueue: jobQueue}
}

// Register creates a new user and returns token pair.
func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest, meta SessionMeta) (string, string, *models.User, error) {
	db := s.db.WithContext(ctx)

	var existingUser models.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return "", "", nil, utils.ErrConflict(constants.ErrEmailAlreadyExists)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", nil, utils.ErrInternal(err)
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return "", "", nil, utils.ErrInternal(err)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Role:         constants.RoleUser,
		IsActive:     true,
	}

	if err := db.Create(user).Error; err != nil {
		return "", "", nil, utils.ErrInternal(err)
	}

	accessToken, refreshToken, err := s.generateTokenPair(ctx, user, meta)
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, user, nil
}

// Login authenticates a user and returns token pair.
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, meta SessionMeta) (string, string, *models.User, error) {
	db := s.db.WithContext(ctx)

	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", nil, utils.ErrUnauthorized(constants.ErrInvalidCredentials)
		}
		return "", "", nil, utils.ErrInternal(err)
	}

	if !user.IsActive {
		return "", "", nil, utils.ErrUnauthorized(constants.ErrAccountDeactivated)
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return "", "", nil, utils.ErrUnauthorized(constants.ErrInvalidCredentials)
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := db.Model(&user).Update("last_login_at", now).Error; err != nil {
		utils.Log.WithError(err).Warn("failed to update last_login_at")
	}

	accessToken, refreshToken, err := s.generateTokenPair(ctx, &user, meta)
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, &user, nil
}

func (s *AuthService) generateTokenPair(ctx context.Context, user *models.User, meta SessionMeta) (accessToken, refreshToken string, err error) {
	db := s.db.WithContext(ctx)

	accessToken, err = utils.GenerateToken(
		user.ID,
		user.Email,
		user.Role,
		user.FirstName,
		user.LastName,
		user.Phone,
	)
	if err != nil {
		return "", "", utils.ErrInternal(err)
	}

	rawRefresh, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", utils.ErrInternal(err)
	}

	hashedRefresh := utils.HashRefreshToken(rawRefresh)
	refreshExpiry := config.AppConfig.JWT.RefreshTokenExpiry
	now := time.Now()

	refreshTokenObj := &models.RefreshToken{
		Token:      hashedRefresh,
		UserID:     user.ID,
		ExpiresAt:  now.Add(refreshExpiry),
		Revoked:    false,
		UserAgent:  meta.UserAgent,
		IPAddress:  meta.IPAddress,
		LastUsedAt: now,
	}
	if err := db.Create(refreshTokenObj).Error; err != nil {
		return "", "", utils.ErrInternal(err)
	}

	return accessToken, rawRefresh, nil
}

// RefreshTokens validates a refresh token, rotates it, and returns a new token pair.
func (s *AuthService) RefreshTokens(ctx context.Context, rawRefreshToken string, meta SessionMeta) (newAccessToken, newRawRefreshToken string, err error) {
	db := s.db.WithContext(ctx)
	hashed := utils.HashRefreshToken(rawRefreshToken)
	now := time.Now()

	var storedToken models.RefreshToken
	err = db.Where("token = ? AND revoked = ? AND expires_at > ?", hashed, false, now).
		First(&storedToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.handleRevokedRefreshToken(ctx, hashed, meta)
		}
		return "", "", utils.ErrInternal(err)
	}

	storedToken.Revoked = true
	storedToken.LastUsedAt = now
	if meta.UserAgent != "" {
		storedToken.UserAgent = meta.UserAgent
	}
	if meta.IPAddress != "" {
		storedToken.IPAddress = meta.IPAddress
	}
	if err := db.Save(&storedToken).Error; err != nil {
		return "", "", utils.ErrInternal(err)
	}

	var user models.User
	if err := db.First(&user, storedToken.UserID).Error; err != nil {
		return "", "", utils.ErrInternal(err)
	}

	newAccessToken, err = utils.GenerateToken(
		user.ID, user.Email, user.Role, user.FirstName, user.LastName, user.Phone,
	)
	if err != nil {
		return "", "", err
	}

	newRawRefreshToken, err = utils.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	hashedNew := utils.HashRefreshToken(newRawRefreshToken)
	newTokenRecord := &models.RefreshToken{
		Token:      hashedNew,
		UserID:     user.ID,
		ExpiresAt:  now.Add(config.AppConfig.JWT.RefreshTokenExpiry),
		Revoked:    false,
		UserAgent:  meta.UserAgent,
		IPAddress:  meta.IPAddress,
		LastUsedAt: now,
	}
	if err := db.Create(newTokenRecord).Error; err != nil {
		return "", "", utils.ErrInternal(err)
	}

	return newAccessToken, newRawRefreshToken, nil
}

func (s *AuthService) handleRevokedRefreshToken(ctx context.Context, hashed string, meta SessionMeta) (string, string, error) {
	db := s.db.WithContext(ctx)

	var revokedToken models.RefreshToken
	err := db.Where("token = ? AND revoked = ?", hashed, true).First(&revokedToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", utils.ErrUnauthorized("invalid or expired refresh token")
		}
		return "", "", utils.ErrInternal(err)
	}

	revokedAt := revokedToken.UpdatedAt
	if revokedAt.IsZero() {
		revokedAt = revokedToken.LastUsedAt
	}

	if time.Since(revokedAt) <= refreshTokenReuseGracePeriod {
		return "", "", utils.ErrUnauthorized("refresh token already rotated")
	}

	if err := s.revokeAllUserRefreshTokens(ctx, revokedToken.UserID); err != nil {
		return "", "", err
	}

	utils.Log.WithFields(map[string]interface{}{
		"user_id":    revokedToken.UserID,
		"ip_address": meta.IPAddress,
	}).Warn("refresh token reuse detected; all sessions revoked")

	return "", "", utils.ErrUnauthorized("refresh token reuse detected")
}

func (s *AuthService) revokeAllUserRefreshTokens(ctx context.Context, userID uint) error {
	result := s.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true)
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	return nil
}

// ListSessions returns active refresh-token sessions for a user.
func (s *AuthService) ListSessions(ctx context.Context, userID uint, currentRefreshToken string) ([]dto.SessionResponse, error) {
	db := s.db.WithContext(ctx)

	var tokens []models.RefreshToken
	if err := db.Where("user_id = ? AND revoked = ? AND expires_at > ?", userID, false, time.Now()).
		Order("last_used_at DESC, created_at DESC").
		Find(&tokens).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	currentHash := ""
	if currentRefreshToken != "" {
		currentHash = utils.HashRefreshToken(currentRefreshToken)
	}

	sessions := make([]dto.SessionResponse, 0, len(tokens))
	for _, token := range tokens {
		lastUsed := token.LastUsedAt
		if lastUsed.IsZero() {
			lastUsed = token.CreatedAt
		}

		sessions = append(sessions, dto.SessionResponse{
			ID:         token.ID,
			UserAgent:  token.UserAgent,
			IPAddress:  token.IPAddress,
			LastUsedAt: lastUsed.Format(time.RFC3339),
			CreatedAt:  token.CreatedAt.Format(time.RFC3339),
			IsCurrent:  currentHash != "" && token.Token == currentHash,
		})
	}

	return sessions, nil
}

// RevokeSession revokes a single refresh-token session.
func (s *AuthService) RevokeSession(ctx context.Context, userID, sessionID uint) error {
	result := s.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("id = ? AND user_id = ? AND revoked = ?", sessionID, userID, false).
		Update("revoked", true)
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound("session not found")
	}
	return nil
}

// RevokeOtherSessions revokes all sessions except the current refresh token.
func (s *AuthService) RevokeOtherSessions(ctx context.Context, userID uint, currentRefreshToken string) error {
	query := s.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false)

	if currentRefreshToken != "" {
		currentHash := utils.HashRefreshToken(currentRefreshToken)
		query = query.Where("token <> ?", currentHash)
	}

	if err := query.Update("revoked", true).Error; err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// Logout revokes refresh tokens for the current session or all sessions for a user.
func (s *AuthService) Logout(ctx context.Context, userID uint, req LogoutRequest) error {
	db := s.db.WithContext(ctx)

	if req.RefreshToken != "" {
		hashed := utils.HashRefreshToken(req.RefreshToken)
		result := db.Model(&models.RefreshToken{}).
			Where("token = ? AND revoked = ?", hashed, false).
			Update("revoked", true)
		if result.Error != nil {
			return utils.ErrInternal(result.Error)
		}
		if result.RowsAffected > 0 {
			return nil
		}
	}

	if userID == 0 {
		return nil
	}

	return db.Model(&models.RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true).Error
}

// ChangePassword use can change password with this services
func (s *AuthService) ChangePassword(ctx context.Context, userID uint, req dto.ChangePasswordRequest) error {
	db := s.db.WithContext(ctx)

	var user models.User
	if err := db.Select("id", "password_hash").First(&user, userID).Error; err != nil {
		return utils.ErrNotFound("user not found")
	}

	if !utils.CheckPasswordHash(req.CurrentPassword, user.PasswordHash) {
		return utils.ErrUnauthorized("current password is incorrect")
	}

	hashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return utils.ErrInternal(err)
	}
	user.PasswordHash = hashed
	if err := db.Save(&user).Error; err != nil {
		return utils.ErrInternal(err)
	}

	return nil
}

// ForgotPassword generates a reset token and sends email.
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	db := s.db.WithContext(ctx)

	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil
	}

	db.Where("user_id = ? AND used_at IS NULL", user.ID).Delete(&models.PasswordResetToken{})

	token, err := utils.GenerateRandomToken()
	if err != nil {
		return utils.ErrInternal(err)
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	resetToken := models.PasswordResetToken{
		UserID:    user.ID,
		Token:     utils.HashRefreshToken(token),
		ExpiresAt: expiresAt,
	}
	if err := db.Create(&resetToken).Error; err != nil {
		return utils.ErrInternal(err)
	}

	s.enqueueSendPasswordResetEmail(user.Email, token)
	return nil
}

func (s *AuthService) enqueueSendPasswordResetEmail(email, token string) {
	if s.jobQueue == nil {
		go utils.SendPasswordResetEmail(email, token)
		return
	}
	frontendURL := ""
	if s.cfg != nil {
		frontendURL = s.cfg.Email.FrontendURL
	}
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)
	subject := "Password Reset Request"
	body := fmt.Sprintf(`<h2>Password Reset</h2><p>Click the link below to reset your password:</p><a href="%s">%s</a><p>Expires in 1 hour.</p>`, resetURL, resetURL)
	if err := s.jobQueue.EnqueueSendEmail(context.Background(), email, subject, body); err != nil {
		utils.Log.WithError(err).Warn("failed to enqueue password reset email; sending inline")
		go utils.SendPasswordResetEmail(email, token)
	}
}

// SendVerificationEmail creates a token and sends verification link.
func (s *AuthService) SendVerificationEmail(ctx context.Context, userID uint) error {
	db := s.db.WithContext(ctx)

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		return utils.ErrNotFound("user not found")
	}
	if user.EmailVerifiedAt != nil {
		return utils.ErrBadRequest("email already verified")
	}

	db.Where("user_id = ? AND used_at IS NULL", userID).Delete(&models.EmailVerificationToken{})

	token, err := utils.GenerateRandomToken()
	if err != nil {
		return utils.ErrInternal(err)
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	vt := models.EmailVerificationToken{
		UserID:    userID,
		Token:     utils.HashRefreshToken(token),
		ExpiresAt: expiresAt,
	}
	if err := db.Create(&vt).Error; err != nil {
		return utils.ErrInternal(err)
	}

	s.enqueueSendVerificationEmail(user.Email, token)
	return nil
}

func (s *AuthService) enqueueSendVerificationEmail(email, token string) {
	if s.jobQueue == nil {
		go utils.SendVerificationEmail(email, token)
		return
	}
	frontendURL := ""
	if s.cfg != nil {
		frontendURL = s.cfg.Email.FrontendURL
	}
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, token)
	subject := "Verify Your Email Address"
	body := fmt.Sprintf(`<h2>Email Verification</h2><p>Please verify your email by clicking below:</p><a href="%s">%s</a><p>Expires in 24 hours.</p>`, verifyURL, verifyURL)
	if err := s.jobQueue.EnqueueSendEmail(context.Background(), email, subject, body); err != nil {
		utils.Log.WithError(err).Warn("failed to enqueue verification email; sending inline")
		go utils.SendVerificationEmail(email, token)
	}
}

// VerifyEmail marks email as verified.
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	db := s.db.WithContext(ctx)

	var vt models.EmailVerificationToken
	err := db.Where("token = ? AND used_at IS NULL AND expires_at > ?", utils.HashRefreshToken(token), time.Now()).
		First(&vt).Error
	if err != nil {
		return utils.ErrBadRequest("invalid or expired verification token")
	}

	var user models.User
	if err := db.First(&user, vt.UserID).Error; err != nil {
		return utils.ErrInternal(err)
	}

	now := time.Now()
	user.EmailVerifiedAt = &now
	if err := db.Save(&user).Error; err != nil {
		return utils.ErrInternal(err)
	}

	vt.UsedAt = &now
	db.Save(&vt)
	return nil
}

// ResetPassword uses a valid reset token to set a new password.
func (s *AuthService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	db := s.db.WithContext(ctx)
	now := time.Now()

	var resetToken models.PasswordResetToken
	err := db.Where("token = ? AND used_at IS NULL AND expires_at > ?", utils.HashRefreshToken(token), now).
		First(&resetToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrBadRequest("invalid or expired reset token")
		}
		return utils.ErrInternal(err)
	}

	var user models.User
	if err := db.First(&user, resetToken.UserID).Error; err != nil {
		return utils.ErrInternal(err)
	}

	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return utils.ErrInternal(err)
	}

	user.PasswordHash = hashed
	if err := db.Save(&user).Error; err != nil {
		return utils.ErrInternal(err)
	}

	resetToken.UsedAt = &now
	if err := db.Save(&resetToken).Error; err != nil {
		utils.Log.WithError(err).Warn("failed to mark reset token as used")
	}

	db.Model(&models.RefreshToken{}).Where("user_id = ?", user.ID).Update("revoked", true)

	return nil
}
