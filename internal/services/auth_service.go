package services

import (
	"context"

	appauth "github.com/alireza-akbarzadeh/luxe/internal/application/auth"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

type SessionMeta = appauth.SessionMeta

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
	app *appauth.Service
}

func NewAuthServices(
	db *gorm.DB,
	cfg *config.Config,
	jobQueue asynq.JobQueue,
	engine *workflow.Engine,
	settings SettingServiceInterface,
) *AuthService {
	repo := postgres.NewAuthRepository(db)
	var legalReader appauth.LegalSettingReader
	if settings != nil {
		legalReader = settingsLegalAdapter{settings: settings}
	}
	return &AuthService{
		app: appauth.NewService(repo, cfg, jobQueue, engine, legalReader),
	}
}

type settingsLegalAdapter struct {
	settings SettingServiceInterface
}

func (a settingsLegalAdapter) GetSettingJSON(ctx context.Context, key string) ([]byte, error) {
	setting, err := a.settings.Get(ctx, key)
	if err != nil || setting == nil {
		return nil, err
	}
	return setting.Value, nil
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest, meta SessionMeta) (string, string, *models.User, error) {
	return s.app.Register(ctx, req, meta)
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, meta SessionMeta) (string, string, *models.User, error) {
	return s.app.Login(ctx, req, meta)
}

func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string, meta SessionMeta) (string, string, error) {
	return s.app.RefreshTokens(ctx, refreshToken, meta)
}

func (s *AuthService) ListSessions(ctx context.Context, userID uint, currentRefreshToken string) ([]dto.SessionResponse, error) {
	return s.app.ListSessions(ctx, userID, currentRefreshToken)
}

func (s *AuthService) RevokeSession(ctx context.Context, userID, sessionID uint) error {
	return s.app.RevokeSession(ctx, userID, sessionID)
}

func (s *AuthService) RevokeOtherSessions(ctx context.Context, userID uint, currentRefreshToken string) error {
	return s.app.RevokeOtherSessions(ctx, userID, currentRefreshToken)
}

func (s *AuthService) Logout(ctx context.Context, userID uint, req LogoutRequest) error {
	return s.app.Logout(ctx, userID, appauth.LogoutRequest{RefreshToken: req.RefreshToken})
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uint, req dto.ChangePasswordRequest) error {
	return s.app.ChangePassword(ctx, userID, req)
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	return s.app.ForgotPassword(ctx, email)
}

func (s *AuthService) SendVerificationEmail(ctx context.Context, userID uint) error {
	return s.app.SendVerificationEmail(ctx, userID)
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	return s.app.VerifyEmail(ctx, token)
}

func (s *AuthService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	return s.app.ResetPassword(ctx, token, newPassword)
}
