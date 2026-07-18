package postgres

import (
	"context"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// UserRepository implements user persistence with GORM.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a GORM-backed user repository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID loads a user model by ID.
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ListUsers returns paginated users matching filters.
func (r *UserRepository) ListUsers(ctx context.Context, filter UserListFilter) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.WithContext(ctx).Model(&models.User{})

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Email != "" {
		query = query.Where("LOWER(email) LIKE LOWER(?)", "%"+filter.Email+"%")
	}
	if filter.Phone != "" {
		query = query.Where("phone LIKE ?", "%"+filter.Phone+"%")
	}
	if filter.FirstName != "" {
		query = query.Where("LOWER(first_name) LIKE LOWER(?)", "%"+filter.FirstName+"%")
	}
	if filter.LastName != "" {
		query = query.Where("LOWER(last_name) LIKE LOWER(?)", "%"+filter.LastName+"%")
	}
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(filter.Limit).
		Offset(filter.Offset).
		Order("created_at DESC, id DESC").
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// Save persists user changes.
func (r *UserRepository) SaveUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// SearchGiftRecipients finds active users for gifting.
// Empty query returns the first `limit` active members (excluding the actor).
// Non-empty query filters by email or phone (partial match).
func (r *UserRepository) SearchGiftRecipients(
	ctx context.Context,
	excludeUserID uint,
	query string,
	limit int,
) ([]models.User, error) {
	query = strings.TrimSpace(query)
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	q := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("is_active = ?", true).
		Where("id != ?", excludeUserID)

	if query != "" {
		if strings.Contains(query, "@") {
			q = q.Where("LOWER(email) LIKE LOWER(?)", "%"+query+"%")
		} else {
			normalizedPhone := strings.TrimPrefix(query, "+")
			q = q.Where(
				"phone LIKE ? OR LOWER(email) LIKE LOWER(?)",
				"%"+normalizedPhone+"%",
				"%"+query+"%",
			)
		}
	}

	var users []models.User
	if err := q.Limit(limit).Order("first_name ASC, last_name ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// DeleteUser soft-deletes a user by ID.
func (r *UserRepository) DeleteUser(ctx context.Context, userID uint) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&models.Product{}, userID)
	return result.RowsAffected, result.Error
}

// UserListFilter holds list query parameters.
type UserListFilter struct {
	Limit     int
	Offset    int
	IsActive  *bool
	Email     string
	Phone     string
	FirstName string
	LastName  string
	Role      string
}
