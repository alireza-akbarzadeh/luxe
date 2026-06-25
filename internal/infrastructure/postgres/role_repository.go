package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// RoleRepository persists roles and permissions with GORM.
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a GORM-backed role repository.
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// ListRoles returns all roles ordered by system flag and name.
func (r *RoleRepository) ListRoles(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.WithContext(ctx).Order("is_system DESC, name ASC").Find(&roles).Error
	return roles, err
}

// FindByID loads a role by primary key.
func (r *RoleRepository) FindByID(ctx context.Context, id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// FindByIDWithPermissions loads a role with permissions preloaded.
func (r *RoleRepository) FindByIDWithPermissions(ctx context.Context, id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// CountBySlug counts roles matching a slug.
func (r *RoleRepository) CountBySlug(ctx context.Context, slug string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Role{}).Where("slug = ?", slug).Count(&count).Error
	return count, err
}

// Create inserts a new role.
func (r *RoleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// Save persists role changes.
func (r *RoleRepository) Save(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete removes a role row.
func (r *RoleRepository) Delete(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Delete(role).Error
}

// CountUsersByRoleSlug counts users assigned to a role slug.
func (r *RoleRepository) CountUsersByRoleSlug(ctx context.Context, slug string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("role = ?", slug).Count(&count).Error
	return count, err
}

// ListPermissions returns all permissions ordered by module and key.
func (r *RoleRepository) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.WithContext(ctx).Order("module ASC, key ASC").Find(&permissions).Error
	return permissions, err
}

// FindPermissionsByIDs loads permissions by id list.
func (r *RoleRepository) FindPermissionsByIDs(ctx context.Context, ids []uint) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&permissions).Error
	return permissions, err
}

// ReplaceRolePermissions replaces role permission associations in a transaction.
func (r *RoleRepository) ReplaceRolePermissions(ctx context.Context, role *models.Role, permissions []models.Permission) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", role.ID).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}
		if len(permissions) == 0 {
			return nil
		}
		return tx.Model(role).Association("Permissions").Replace(permissions)
	})
}

// CountPermissionByRoleAndKey checks if a role slug has a permission key.
func (r *RoleRepository) CountPermissionByRoleAndKey(ctx context.Context, roleSlug, permissionKey string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.RolePermission{}).
		Joins("JOIN roles r ON r.id = role_permissions.role_id").
		Joins("JOIN permissions p ON p.id = role_permissions.permission_id").
		Where("r.slug = ? AND p.key = ?", roleSlug, permissionKey).
		Count(&count).Error
	return count, err
}

// CountRolePermissions counts permission rows linked to a role.
func (r *RoleRepository) CountRolePermissions(ctx context.Context, roleID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.RolePermission{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}
