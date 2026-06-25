package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

const refreshTokenReuseGracePeriod = 30 * time.Second

// SessionMeta carries client metadata for auth sessions.
type SessionMeta struct {
	UserAgent string
	IPAddress string
}

// LegalSettingReader resolves legal document version strings from settings.
type LegalSettingReader interface {
	GetSettingJSON(ctx context.Context, key string) ([]byte, error)
}

// Service orchestrates authentication use cases.
type Service struct {
	repo     *postgres.AuthRepository
	cfg      *config.Config
	jobQueue asynq.JobQueue
	engine   *workflow.Engine
	settings LegalSettingReader
}

// NewService creates auth use cases.
func NewService(
	repo *postgres.AuthRepository,
	cfg *config.Config,
	jobQueue asynq.JobQueue,
	engine *workflow.Engine,
	settings LegalSettingReader,
) *Service {
	return &Service{repo: repo, cfg: cfg, jobQueue: jobQueue, engine: engine, settings: settings}
}

type legalDocumentMeta struct {
	Version string `json:"version"`
}

func legalVersionFromSetting(ctx context.Context, settings LegalSettingReader, key string) string {
	if settings == nil {
		return "unknown"
	}
	raw, err := settings.GetSettingJSON(ctx, key)
	if err != nil || len(raw) == 0 {
		return "unknown"
	}
	var meta legalDocumentMeta
	if err := json.Unmarshal(raw, &meta); err != nil || meta.Version == "" {
		return "unknown"
	}
	return meta.Version
}

// Register creates a new user and returns token pair.
func (s *Service) Register(ctx context.Context, req dto.RegisterRequest, meta SessionMeta) (string, string, *models.User, error) {
	if !req.AcceptTerms || !req.AcceptPrivacy {
		return "", "", nil, utils.ErrBadRequest(constants.ErrLegalAcceptanceRequired)
	}

	_, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err == nil {
		return "", "", nil, utils.ErrConflict(constants.ErrEmailAlreadyExists)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", nil, utils.ErrInternal(err)
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return "", "", nil, utils.ErrInternal(err)
	}

	now := time.Now()
	termsVersion := legalVersionFromSetting(ctx, s.settings, constants.SettingKeyLegalTerms)
	privacyVersion := legalVersionFromSetting(ctx, s.settings, constants.SettingKeyLegalPrivacy)

	user := &models.User{
		Email:             req.Email,
		PasswordHash:      hashedPassword,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Phone:             req.Phone,
		Role:              constants.RoleUser,
		IsActive:          true,
		TermsAcceptedAt:   &now,
		PrivacyAcceptedAt: &now,
		TermsVersion:      &termsVersion,
		PrivacyVersion:    &privacyVersion,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return "", "", nil, utils.ErrInternal(err)
	}

	appworkflow.SyncUserState(ctx, s.engine, user.ID, "email_verification_pending", "register", nil)

	accessToken, refreshToken, err := s.generateTokenPair(ctx, user, meta)
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, user, nil
}

// Login authenticates a user and returns token pair.
func (s *Service) Login(ctx context.Context, req dto.LoginRequest, meta SessionMeta) (string, string, *models.User, error) {
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
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
	if err := s.repo.UpdateUserLastLogin(ctx, user.ID, now); err != nil {
		utils.Log.WithError(err).Warn("failed to update last_login_at")
	}
	user.LastLoginAt = &now

	accessToken, refreshToken, err := s.generateTokenPair(ctx, user, meta)
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, user, nil
}

func (s *Service) generateTokenPair(ctx context.Context, user *models.User, meta SessionMeta) (accessToken, refreshToken string, err error) {
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
	if err := s.repo.CreateRefreshToken(ctx, refreshTokenObj); err != nil {
		return "", "", utils.ErrInternal(err)
	}

	return accessToken, rawRefresh, nil
}

// RefreshTokens validates a refresh token, rotates it, and returns a new token pair.
func (s *Service) RefreshTokens(ctx context.Context, rawRefreshToken string, meta SessionMeta) (newAccessToken, newRawRefreshToken string, err error) {
	hashed := utils.HashRefreshToken(rawRefreshToken)
	now := time.Now()

	storedToken, err := s.repo.FindRefreshTokenByHash(ctx, hashed, now)
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
	if err := s.repo.SaveRefreshToken(ctx, storedToken); err != nil {
		return "", "", utils.ErrInternal(err)
	}

	user, err := s.repo.FindUserByID(ctx, storedToken.UserID)
	if err != nil {
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
	if err := s.repo.CreateRefreshToken(ctx, newTokenRecord); err != nil {
		return "", "", utils.ErrInternal(err)
	}

	return newAccessToken, newRawRefreshToken, nil
}

func (s *Service) handleRevokedRefreshToken(ctx context.Context, hashed string, meta SessionMeta) (string, string, error) {
	revokedToken, err := s.repo.FindRevokedRefreshTokenByHash(ctx, hashed)
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

	if err := s.repo.RevokeAllUserRefreshTokens(ctx, revokedToken.UserID); err != nil {
		return "", "", utils.ErrInternal(err)
	}

	utils.Log.WithFields(map[string]interface{}{
		"user_id":    revokedToken.UserID,
		"ip_address": meta.IPAddress,
	}).Warn("refresh token reuse detected; all sessions revoked")

	return "", "", utils.ErrUnauthorized("refresh token reuse detected")
}

// ListSessions returns active refresh-token sessions for a user.
func (s *Service) ListSessions(ctx context.Context, userID uint, currentRefreshToken string) ([]dto.SessionResponse, error) {
	tokens, err := s.repo.ListActiveRefreshTokens(ctx, userID, time.Now())
	if err != nil {
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
func (s *Service) RevokeSession(ctx context.Context, userID, sessionID uint) error {
	rows, err := s.repo.RevokeSession(ctx, userID, sessionID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("session not found")
	}
	return nil
}

// RevokeOtherSessions revokes all sessions except the current refresh token.
func (s *Service) RevokeOtherSessions(ctx context.Context, userID uint, currentRefreshToken string) error {
	currentHash := ""
	if currentRefreshToken != "" {
		currentHash = utils.HashRefreshToken(currentRefreshToken)
	}
	if err := s.repo.RevokeOtherSessions(ctx, userID, currentHash); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// LogoutRequest is the logout command payload.
type LogoutRequest struct {
	RefreshToken string
}

// Logout revokes refresh tokens for the current session or all sessions for a user.
func (s *Service) Logout(ctx context.Context, userID uint, req LogoutRequest) error {
	if req.RefreshToken != "" {
		hashed := utils.HashRefreshToken(req.RefreshToken)
		rows, err := s.repo.RevokeRefreshTokenByHash(ctx, hashed)
		if err != nil {
			return utils.ErrInternal(err)
		}
		if rows > 0 {
			return nil
		}
	}

	if userID == 0 {
		return nil
	}

	return s.repo.RevokeAllUserTokens(ctx, userID)
}

// ChangePassword updates the user's password.
func (s *Service) ChangePassword(ctx context.Context, userID uint, req dto.ChangePasswordRequest) error {
	user, err := s.repo.FindUserPasswordHash(ctx, userID)
	if err != nil {
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
	if err := s.repo.SaveUser(ctx, user); err != nil {
		return utils.ErrInternal(err)
	}

	return nil
}

// ForgotPassword generates a reset token and sends email.
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil
	}

	if err := s.repo.DeleteUnusedPasswordResetTokens(ctx, user.ID); err != nil {
		return utils.ErrInternal(err)
	}

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
	if err := s.repo.CreatePasswordResetToken(ctx, &resetToken); err != nil {
		return utils.ErrInternal(err)
	}

	s.enqueueSendPasswordResetEmail(user.Email, token)
	return nil
}

func (s *Service) enqueueSendPasswordResetEmail(email, token string) {
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
func (s *Service) SendVerificationEmail(ctx context.Context, userID uint) error {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return utils.ErrNotFound("user not found")
	}
	if user.EmailVerifiedAt != nil {
		return utils.ErrBadRequest("email already verified")
	}

	if err := s.repo.DeleteUnusedEmailVerificationTokens(ctx, userID); err != nil {
		return utils.ErrInternal(err)
	}

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
	if err := s.repo.CreateEmailVerificationToken(ctx, &vt); err != nil {
		return utils.ErrInternal(err)
	}

	s.enqueueSendVerificationEmail(user.Email, token)
	return nil
}

func (s *Service) enqueueSendVerificationEmail(email, token string) {
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
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	vt, err := s.repo.FindEmailVerificationToken(ctx, utils.HashRefreshToken(token), time.Now())
	if err != nil {
		return utils.ErrBadRequest("invalid or expired verification token")
	}

	user, err := s.repo.FindUserByID(ctx, vt.UserID)
	if err != nil {
		return utils.ErrInternal(err)
	}

	now := time.Now()
	user.EmailVerifiedAt = &now
	if err := s.repo.SaveUser(ctx, user); err != nil {
		return utils.ErrInternal(err)
	}

	vt.UsedAt = &now
	_ = s.repo.SaveEmailVerificationToken(ctx, vt)

	if err := appworkflow.ApplyEvent(ctx, s.engine, workflow.TransitionRequest{
		WorkflowKey: constants.WorkflowEntityUser,
		EntityID:    user.ID,
		Event:       "verify_email",
		ActorID:     &user.ID,
		ActorRole:   user.Role,
	}); err != nil {
		appworkflow.SyncUserState(ctx, s.engine, user.ID, "active", "verify_email", &user.ID)
	}

	return nil
}

// ResetPassword uses a valid reset token to set a new password.
func (s *Service) ResetPassword(ctx context.Context, token string, newPassword string) error {
	now := time.Now()

	resetToken, err := s.repo.FindPasswordResetToken(ctx, utils.HashRefreshToken(token), now)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrBadRequest("invalid or expired reset token")
		}
		return utils.ErrInternal(err)
	}

	user, err := s.repo.FindUserByID(ctx, resetToken.UserID)
	if err != nil {
		return utils.ErrInternal(err)
	}

	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return utils.ErrInternal(err)
	}

	user.PasswordHash = hashed
	if err := s.repo.SaveUser(ctx, user); err != nil {
		return utils.ErrInternal(err)
	}

	resetToken.UsedAt = &now
	if err := s.repo.SavePasswordResetToken(ctx, resetToken); err != nil {
		utils.Log.WithError(err).Warn("failed to mark reset token as used")
	}

	_ = s.repo.RevokeAllUserTokens(ctx, user.ID)

	return nil
}
