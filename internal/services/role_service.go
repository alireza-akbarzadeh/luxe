package services

import (
	"context"
	"errors"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

type RoleServiceInterface interface {
	ListRoles(ctx context.Context) ([]dto.RoleResponse, error)
	GetRole(ctx context.Context, id uint) (*dto.RoleResponse, error)
	CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error)
	UpdateRole(ctx context.Context, id uint, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	DeleteRole(ctx context.Context, id uint) error
	ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error)
	SetRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint) (*dto.RoleResponse, error)
	RoleSlugExists(ctx context.Context, slug string) (bool, error)
}

type roleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) RoleServiceInterface {
	return &roleService{db: db}
}

func (s *roleService) ListRoles(ctx context.Context) ([]dto.RoleResponse, error) {
	var roles []models.Role
	if err := s.db.WithContext(ctx).Order("is_system DESC, name ASC").Find(&roles).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	result := make([]dto.RoleResponse, 0, len(roles))
	for _, role := range roles {
		item, err := s.toRoleSummary(ctx, &role)
		if err != nil {
			return nil, err
		}
		result = append(result, *item)
	}
	return result, nil
}

func (s *roleService) GetRole(ctx context.Context, id uint) (*dto.RoleResponse, error) {
	var role models.Role
	if err := s.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("role not found")
		}
		return nil, utils.ErrInternal(err)
	}

	resp, err := s.toRoleSummary(ctx, &role)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(role.Permissions))
	ids := make([]uint, 0, len(role.Permissions))
	for _, permission := range role.Permissions {
		keys = append(keys, permission.Key)
		ids = append(ids, permission.ID)
	}
	resp.PermissionKeys = keys
	resp.PermissionIDs = ids
	return resp, nil
}

func (s *roleService) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	if slug == "" {
		return nil, utils.ErrBadRequest("slug is required")
	}

	var existing int64
	if err := s.db.WithContext(ctx).Model(&models.Role{}).Where("slug = ?", slug).Count(&existing).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	if existing > 0 {
		return nil, utils.ErrBadRequest("role slug already exists")
	}

	var description *string
	if trimmed := strings.TrimSpace(req.Description); trimmed != "" {
		description = &trimmed
	}

	role := models.Role{
		Name:        strings.TrimSpace(req.Name),
		Slug:        slug,
		Description: description,
		IsSystem:    false,
	}
	if err := s.db.WithContext(ctx).Create(&role).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return s.GetRole(ctx, role.ID)
}

func (s *roleService) UpdateRole(ctx context.Context, id uint, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	var role models.Role
	if err := s.db.WithContext(ctx).First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("role not found")
		}
		return nil, utils.ErrInternal(err)
	}

	var description *string
	if trimmed := strings.TrimSpace(req.Description); trimmed != "" {
		description = &trimmed
	}

	role.Name = strings.TrimSpace(req.Name)
	role.Description = description
	if err := s.db.WithContext(ctx).Save(&role).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return s.GetRole(ctx, role.ID)
}

func (s *roleService) DeleteRole(ctx context.Context, id uint) error {
	var role models.Role
	if err := s.db.WithContext(ctx).First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("role not found")
		}
		return utils.ErrInternal(err)
	}
	if role.IsSystem {
		return utils.ErrBadRequest("system roles cannot be deleted")
	}

	var userCount int64
	if err := s.db.WithContext(ctx).Model(&models.User{}).Where("role = ?", role.Slug).Count(&userCount).Error; err != nil {
		return utils.ErrInternal(err)
	}
	if userCount > 0 {
		return utils.ErrBadRequest("role is assigned to users and cannot be deleted")
	}

	if err := s.db.WithContext(ctx).Delete(&role).Error; err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (s *roleService) ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	var permissions []models.Permission
	if err := s.db.WithContext(ctx).Order("module ASC, key ASC").Find(&permissions).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	result := make([]dto.PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		result = append(result, dto.PermissionResponse{
			ID:          permission.ID,
			Key:         permission.Key,
			Module:      permission.Module,
			Description: permission.Description,
		})
	}
	return result, nil
}

func (s *roleService) SetRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint) (*dto.RoleResponse, error) {
	var role models.Role
	if err := s.db.WithContext(ctx).First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("role not found")
		}
		return nil, utils.ErrInternal(err)
	}

	var permissions []models.Permission
	if len(permissionIDs) > 0 {
		if err := s.db.WithContext(ctx).Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
			return nil, utils.ErrInternal(err)
		}
		if len(permissions) != len(permissionIDs) {
			return nil, utils.ErrBadRequest("one or more permissions were not found")
		}
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}
		if len(permissions) == 0 {
			return nil
		}
		return tx.Model(&role).Association("Permissions").Replace(permissions)
	}); err != nil {
		return nil, utils.ErrInternal(err)
	}

	return s.GetRole(ctx, roleID)
}

func (s *roleService) RoleSlugExists(ctx context.Context, slug string) (bool, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Role{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, utils.ErrInternal(err)
	}
	return count > 0, nil
}

func (s *roleService) toRoleSummary(ctx context.Context, role *models.Role) (*dto.RoleResponse, error) {
	var userCount int64
	if err := s.db.WithContext(ctx).Model(&models.User{}).Where("role = ?", role.Slug).Count(&userCount).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var permissionCount int64
	if err := s.db.WithContext(ctx).Model(&models.RolePermission{}).Where("role_id = ?", role.ID).Count(&permissionCount).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	return &dto.RoleResponse{
		ID:              role.ID,
		Name:            role.Name,
		Slug:            role.Slug,
		Description:     role.Description,
		IsSystem:        role.IsSystem,
		UserCount:       userCount,
		PermissionCount: permissionCount,
		CreatedAt:       role.CreatedAt,
		UpdatedAt:       role.UpdatedAt,
	}, nil
}
