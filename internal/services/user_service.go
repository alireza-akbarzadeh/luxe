package services

import (
	"context"

	appuser "github.com/alireza-akbarzadeh/luxe/internal/application/user"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type UserServiceInterface interface {
	GetUserByID(userID uint) (*models.User, error)
	GetUsers(filter UserFilter) ([]models.User, int64, error)
	UpdateUserProfile(userID uint, req UpdateProfileRequest) (*models.User, error)
	DeleteUser(userID uint) error
}

type UserFilter struct {
	Limit     int    `form:"limit" binding:"omitempty,max=100"`
	Offset    int    `form:"offset" binding:"omitempty,min=0"`
	IsActive  *bool  `form:"is_active"`
	Email     string `form:"email"`
	Phone     string `form:"phone"`
	FirstName string `form:"first_name"`
	LastName  string `form:"last_name"`
	Role      string `form:"role" binding:"omitempty,oneof=user admin moderator"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
	Phone     string `json:"phone" validate:"omitempty,e164"`
	Role      string `form:"role" binding:"omitempty,oneof=user admin moderator"`
}

type UserService struct {
	queries  *appuser.Queries
	commands *appuser.Commands
}

func NewUserService(db *gorm.DB, _ *config.Config) *UserService {
	repo := postgres.NewUserRepository(db)
	return &UserService{
		queries:  appuser.NewQueries(repo),
		commands: appuser.NewCommands(repo),
	}
}

func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	return s.queries.GetByID(context.Background(), userID)
}

func (s *UserService) GetUsers(filter UserFilter) ([]models.User, int64, error) {
	return s.queries.List(context.Background(), appuser.UserFilter{
		Limit:     filter.Limit,
		Offset:    filter.Offset,
		IsActive:  filter.IsActive,
		Email:     filter.Email,
		Phone:     filter.Phone,
		FirstName: filter.FirstName,
		LastName:  filter.LastName,
		Role:      filter.Role,
	})
}

func (s *UserService) UpdateUserProfile(userID uint, req UpdateProfileRequest) (*models.User, error) {
	return s.commands.UpdateProfile(context.Background(), userID, appuser.UpdateProfileInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      req.Role,
	})
}

func (s *UserService) DeleteUser(userID uint) error {
	return s.commands.Delete(context.Background(), userID)
}
