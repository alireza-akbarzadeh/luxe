package role

import (
	"context"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Queries orchestrates role read use cases.
type Queries struct {
	repo *postgres.RoleRepository
}

// NewQueries creates role query use cases.
func NewQueries(repo *postgres.RoleRepository) *Queries {
	return &Queries{repo: repo}
}

// ListRoles returns all roles with summary counts.
func (q *Queries) ListRoles(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := q.repo.ListRoles(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	result := make([]dto.RoleResponse, 0, len(roles))
	for i := range roles {
		item, err := q.toRoleSummary(ctx, &roles[i])
		if err != nil {
			return nil, err
		}
		result = append(result, *item)
	}
	return result, nil
}

// GetRole returns a role with permission keys and ids.
func (q *Queries) GetRole(ctx context.Context, id uint) (*dto.RoleResponse, error) {
	role, err := q.repo.FindByIDWithPermissions(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, utils.ErrNotFound("role not found")
		}
		return nil, utils.ErrInternal(err)
	}

	resp, err := q.toRoleSummary(ctx, role)
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

// ListPermissions returns all permissions.
func (q *Queries) ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	permissions, err := q.repo.ListPermissions(ctx)
	if err != nil {
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

// RoleSlugExists reports whether a slug is already taken.
func (q *Queries) RoleSlugExists(ctx context.Context, slug string) (bool, error) {
	count, err := q.repo.CountBySlug(ctx, slug)
	if err != nil {
		return false, utils.ErrInternal(err)
	}
	return count > 0, nil
}

// HasPermission checks whether a role slug grants a permission key.
func (q *Queries) HasPermission(ctx context.Context, roleSlug, permissionKey string) (bool, error) {
	if roleSlug == constants.RoleAdmin {
		return true, nil
	}
	count, err := q.repo.CountPermissionByRoleAndKey(ctx, roleSlug, permissionKey)
	if err != nil {
		return false, utils.ErrInternal(err)
	}
	return count > 0, nil
}

func (q *Queries) toRoleSummary(ctx context.Context, role *models.Role) (*dto.RoleResponse, error) {
	userCount, err := q.repo.CountUsersByRoleSlug(ctx, role.Slug)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	permissionCount, err := q.repo.CountRolePermissions(ctx, role.ID)
	if err != nil {
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

// Commands orchestrates role write use cases.
type Commands struct {
	repo    *postgres.RoleRepository
	queries *Queries
}

// NewCommands creates role command use cases.
func NewCommands(repo *postgres.RoleRepository, queries *Queries) *Commands {
	return &Commands{repo: repo, queries: queries}
}

// CreateRole inserts a new role.
func (c *Commands) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	if slug == "" {
		return nil, utils.ErrBadRequest("slug is required")
	}

	existing, err := c.repo.CountBySlug(ctx, slug)
	if err != nil {
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
	if err := c.repo.Create(ctx, &role); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return c.queries.GetRole(ctx, role.ID)
}

// UpdateRole modifies an existing role.
func (c *Commands) UpdateRole(ctx context.Context, id uint, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	role, err := c.repo.FindByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
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
	if err := c.repo.Save(ctx, role); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return c.queries.GetRole(ctx, role.ID)
}

// DeleteRole removes a non-system role with no assigned users.
func (c *Commands) DeleteRole(ctx context.Context, id uint) error {
	role, err := c.repo.FindByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return utils.ErrNotFound("role not found")
		}
		return utils.ErrInternal(err)
	}
	if role.IsSystem {
		return utils.ErrBadRequest("system roles cannot be deleted")
	}

	userCount, err := c.repo.CountUsersByRoleSlug(ctx, role.Slug)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if userCount > 0 {
		return utils.ErrBadRequest("role is assigned to users and cannot be deleted")
	}

	if err := c.repo.Delete(ctx, role); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// SetRolePermissions replaces permissions assigned to a role.
func (c *Commands) SetRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint) (*dto.RoleResponse, error) {
	role, err := c.repo.FindByID(ctx, roleID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, utils.ErrNotFound("role not found")
		}
		return nil, utils.ErrInternal(err)
	}

	var permissions []models.Permission
	if len(permissionIDs) > 0 {
		permissions, err = c.repo.FindPermissionsByIDs(ctx, permissionIDs)
		if err != nil {
			return nil, utils.ErrInternal(err)
		}
		if len(permissions) != len(permissionIDs) {
			return nil, utils.ErrBadRequest("one or more permissions were not found")
		}
	}

	if err := c.repo.ReplaceRolePermissions(ctx, role, permissions); err != nil {
		return nil, utils.ErrInternal(err)
	}

	return c.queries.GetRole(ctx, roleID)
}
