package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// AuthRepository implements auth persistence with GORM.
type AuthRepository struct {
	db *gorm.DB
}

// NewAuthRepository creates a GORM-backed auth repository.
func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// FindUserByEmail loads a user by email.
func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser inserts a new user.
func (r *AuthRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// UpdateUserLastLogin sets last_login_at for a user.
func (r *AuthRepository) UpdateUserLastLogin(ctx context.Context, userID uint, at time.Time) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("last_login_at", at).Error
}

// FindUserByID loads a user by primary key.
func (r *AuthRepository) FindUserByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserPasswordHash loads id and password_hash for a user.
func (r *AuthRepository) FindUserPasswordHash(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Select("id", "password_hash").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// SaveUser persists user changes.
func (r *AuthRepository) SaveUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// CreateRefreshToken inserts a refresh token row.
func (r *AuthRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// FindRefreshTokenByHash loads an active refresh token by hash.
func (r *AuthRepository) FindRefreshTokenByHash(ctx context.Context, hashed string, now time.Time) (*models.RefreshToken, error) {
	var stored models.RefreshToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND revoked = ? AND expires_at > ?", hashed, false, now).
		First(&stored).Error
	if err != nil {
		return nil, err
	}
	return &stored, nil
}

// FindRevokedRefreshTokenByHash loads a revoked refresh token by hash.
func (r *AuthRepository) FindRevokedRefreshTokenByHash(ctx context.Context, hashed string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.WithContext(ctx).Where("token = ? AND revoked = ?", hashed, true).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// SaveRefreshToken persists refresh token changes.
func (r *AuthRepository) SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

// RevokeAllUserRefreshTokens revokes all active tokens for a user.
func (r *AuthRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true).Error
}

// ListActiveRefreshTokens returns active sessions for a user.
func (r *AuthRepository) ListActiveRefreshTokens(ctx context.Context, userID uint, now time.Time) ([]models.RefreshToken, error) {
	var tokens []models.RefreshToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked = ? AND expires_at > ?", userID, false, now).
		Order("last_used_at DESC, created_at DESC").
		Find(&tokens).Error
	return tokens, err
}

// RevokeSession revokes a single session by ID and user.
func (r *AuthRepository) RevokeSession(ctx context.Context, userID, sessionID uint) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("id = ? AND user_id = ? AND revoked = ?", sessionID, userID, false).
		Update("revoked", true)
	return result.RowsAffected, result.Error
}

// RevokeOtherSessions revokes all sessions except the current token hash.
func (r *AuthRepository) RevokeOtherSessions(ctx context.Context, userID uint, currentHash string) error {
	query := r.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false)
	if currentHash != "" {
		query = query.Where("token <> ?", currentHash)
	}
	return query.Update("revoked", true).Error
}

// RevokeRefreshTokenByHash revokes a token by its hash.
func (r *AuthRepository) RevokeRefreshTokenByHash(ctx context.Context, hashed string) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("token = ? AND revoked = ?", hashed, false).
		Update("revoked", true)
	return result.RowsAffected, result.Error
}

// RevokeAllUserTokens revokes every refresh token for a user.
func (r *AuthRepository) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("user_id = ?", userID).Update("revoked", true).Error
}

// DeleteUnusedPasswordResetTokens removes pending reset tokens for a user.
func (r *AuthRepository) DeleteUnusedPasswordResetTokens(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND used_at IS NULL", userID).
		Delete(&models.PasswordResetToken{}).Error
}

// CreatePasswordResetToken inserts a password reset token.
func (r *AuthRepository) CreatePasswordResetToken(ctx context.Context, token *models.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// DeleteUnusedEmailVerificationTokens removes pending verification tokens.
func (r *AuthRepository) DeleteUnusedEmailVerificationTokens(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND used_at IS NULL", userID).
		Delete(&models.EmailVerificationToken{}).Error
}

// CreateEmailVerificationToken inserts an email verification token.
func (r *AuthRepository) CreateEmailVerificationToken(ctx context.Context, token *models.EmailVerificationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// FindEmailVerificationToken loads a valid verification token.
func (r *AuthRepository) FindEmailVerificationToken(ctx context.Context, hashed string, now time.Time) (*models.EmailVerificationToken, error) {
	var vt models.EmailVerificationToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND used_at IS NULL AND expires_at > ?", hashed, now).
		First(&vt).Error
	if err != nil {
		return nil, err
	}
	return &vt, nil
}

// SaveEmailVerificationToken persists verification token changes.
func (r *AuthRepository) SaveEmailVerificationToken(ctx context.Context, vt *models.EmailVerificationToken) error {
	return r.db.WithContext(ctx).Save(vt).Error
}

// FindPasswordResetToken loads a valid reset token.
func (r *AuthRepository) FindPasswordResetToken(ctx context.Context, hashed string, now time.Time) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND used_at IS NULL AND expires_at > ?", hashed, now).
		First(&resetToken).Error
	if err != nil {
		return nil, err
	}
	return &resetToken, nil
}

// SavePasswordResetToken persists reset token changes.
func (r *AuthRepository) SavePasswordResetToken(ctx context.Context, token *models.PasswordResetToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

// FindUserByPhone loads a user by exact phone match.
func (r *AuthRepository) FindUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteUnusedLoginOTPs removes unused login OTP rows for a user.
func (r *AuthRepository) DeleteUnusedLoginOTPs(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND used_at IS NULL", userID).
		Delete(&models.LoginOTP{}).Error
}

// CreateLoginOTP inserts a login OTP row.
func (r *AuthRepository) CreateLoginOTP(ctx context.Context, otp *models.LoginOTP) error {
	return r.db.WithContext(ctx).Create(otp).Error
}

// FindValidLoginOTP loads the latest unused, unexpired OTP for a user.
func (r *AuthRepository) FindValidLoginOTP(ctx context.Context, userID uint, now time.Time) (*models.LoginOTP, error) {
	var otp models.LoginOTP
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND used_at IS NULL AND expires_at > ?", userID, now).
		Order("created_at DESC").
		First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

// SaveLoginOTP persists login OTP changes.
func (r *AuthRepository) SaveLoginOTP(ctx context.Context, otp *models.LoginOTP) error {
	return r.db.WithContext(ctx).Save(otp).Error
}
