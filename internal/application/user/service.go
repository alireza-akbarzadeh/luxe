package user

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// UserFilter defines filter and pagination parameters for listing users.
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

// UpdateProfileRequest is the HTTP payload for profile updates.
type UpdateProfileRequest struct {
	FirstName string  `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string  `json:"last_name" validate:"required,min=1,max=100"`
	Phone     string  `json:"phone" validate:"omitempty,e164"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,max=2048,url"`
	Role      string  `form:"role" binding:"omitempty,oneof=user admin moderator"`
}

// UpdateProfileInput updates non-sensitive user fields.
type UpdateProfileInput struct {
	FirstName string
	LastName  string
	Phone     string
	AvatarURL *string
	Role      string
}

// Queries handles user read use cases.
type Queries struct {
	repo *postgres.UserRepository
}

// NewQueries creates user query use cases.
func NewQueries(repo *postgres.UserRepository) *Queries {
	return &Queries{repo: repo}
}

// GetByID retrieves a user by ID.
func (q *Queries) GetByID(_ context.Context, userID uint) (*models.User, error) {
	user, err := q.repo.FindByID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound()
		}
		return nil, utils.ErrInternal(err)
	}
	return user, nil
}

// List returns paginated users matching filters.
func (q *Queries) List(_ context.Context, filter UserFilter) ([]models.User, int64, error) {
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	users, total, err := q.repo.ListUsers(context.Background(), postgres.UserListFilter{
		Limit:     filter.Limit,
		Offset:    filter.Offset,
		IsActive:  filter.IsActive,
		Email:     filter.Email,
		Phone:     filter.Phone,
		FirstName: filter.FirstName,
		LastName:  filter.LastName,
		Role:      filter.Role,
	})
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return users, total, nil
}

// Commands handles user write use cases.
type Commands struct {
	repo *postgres.UserRepository
}

// NewCommands creates user command use cases.
func NewCommands(repo *postgres.UserRepository) *Commands {
	return &Commands{repo: repo}
}

// UpdateProfile updates non-sensitive user fields.
func (c *Commands) UpdateProfile(_ context.Context, userID uint, in UpdateProfileInput) (*models.User, error) {
	user, err := c.repo.FindByID(context.Background(), userID)
	if err != nil {
		return nil, utils.ErrNotFound()
	}

	user.FirstName = in.FirstName
	user.LastName = in.LastName
	user.Phone = in.Phone
	if in.AvatarURL != nil {
		user.AvatarURL = *in.AvatarURL
	}
	user.Role = in.Role

	if err := c.repo.SaveUser(context.Background(), user); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return user, nil
}

// Delete soft-deletes a user (legacy behavior preserved).
func (c *Commands) Delete(_ context.Context, userID uint) error {
	rows, err := c.repo.DeleteUser(context.Background(), userID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("product not found")
	}
	return nil
}
