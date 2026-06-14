package services

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// uint test

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)
	return gormDB, mock
}

// uint test for the user login
func TestAuthService_Login(t *testing.T) {
	hashed, _ := utils.HashPassword("password123")
	user := models.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: hashed,
		IsActive:     true,
		Role:         "user",
	}
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:             "testsecret",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 168 * time.Hour,
		},
	}

	config.AppConfig = cfg

	t.Run("success login", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		svc := NewAuthServices(gormDB, cfg)

		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "is_active", "role", "last_login_at"}).
			AddRow(user.ID, user.Email, user.PasswordHash, user.IsActive, user.Role, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1`) + `.*`).
			WithArgs("test@example.com", sqlmock.AnyArg()).
			WillReturnRows(rows)

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET`)+`.*`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), user.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "refresh_tokens"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		accessToken, refreshToken, returnedUser, err := svc.Login(dto.LoginRequest{
			Email: "test@example.com", Password: "password123",
		}, SessionMeta{})

		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.NotNil(t, returnedUser)
		assert.Equal(t, user.ID, returnedUser.ID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
	// user not found
	t.Run("user not found", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		svc := NewAuthServices(gormDB, cfg)
		mock.ExpectQuery(`SELECT \* FROM "users"`).WillReturnError(gorm.ErrRecordNotFound)

		accessToken, refreshToken, user, err := svc.Login(dto.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "anything",
		}, SessionMeta{})

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid email or password")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid password", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		svc := NewAuthServices(gormDB, cfg)

		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "is_active", "role", "last_login_at"}).AddRow(user.ID, user.Email, hashed, true, "user", nil)
		mock.ExpectQuery(`SELECT \* FROM "users"`).WillReturnRows(rows)

		accessToken, refreshToken, user, err := svc.Login(dto.LoginRequest{
			Email:    "test@example.con",
			Password: "123456",
		}, SessionMeta{})

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid email or password")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("inactive user", func(t *testing.T) {
		gormDB, mock := setupTestDB(t)
		svc := NewAuthServices(gormDB, cfg)
		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "is_active", "role", "last_login_at"}).
			AddRow(user.ID, user.Email, user.PasswordHash, false, user.Role, nil) // Set is_active to false here!
		mock.ExpectQuery(`SELECT \* FROM "users"`).WillReturnRows(rows)

		accessToken, refreshToken, user, err := svc.Login(dto.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}, SessionMeta{})
		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "account is deactivated")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
