package services

import (
	"context"

	approle "github.com/alireza-akbarzadeh/luxe/internal/application/role"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
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
	HasPermission(ctx context.Context, roleSlug, permissionKey string) (bool, error)
}

type roleService struct {
	commands *approle.Commands
	queries  *approle.Queries
}

func NewRoleService(db *gorm.DB) RoleServiceInterface {
	repo := postgres.NewRoleRepository(db)
	queries := approle.NewQueries(repo)
	return &roleService{
		commands: approle.NewCommands(repo, queries),
		queries:  queries,
	}
}

func (s *roleService) ListRoles(ctx context.Context) ([]dto.RoleResponse, error) {
	return s.queries.ListRoles(ctx)
}

func (s *roleService) GetRole(ctx context.Context, id uint) (*dto.RoleResponse, error) {
	return s.queries.GetRole(ctx, id)
}

func (s *roleService) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	return s.commands.CreateRole(ctx, req)
}

func (s *roleService) UpdateRole(ctx context.Context, id uint, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	return s.commands.UpdateRole(ctx, id, req)
}

func (s *roleService) DeleteRole(ctx context.Context, id uint) error {
	return s.commands.DeleteRole(ctx, id)
}

func (s *roleService) ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	return s.queries.ListPermissions(ctx)
}

func (s *roleService) SetRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint) (*dto.RoleResponse, error) {
	return s.commands.SetRolePermissions(ctx, roleID, permissionIDs)
}

func (s *roleService) RoleSlugExists(ctx context.Context, slug string) (bool, error) {
	return s.queries.RoleSlugExists(ctx, slug)
}

func (s *roleService) HasPermission(ctx context.Context, roleSlug, permissionKey string) (bool, error) {
	return s.queries.HasPermission(ctx, roleSlug, permissionKey)
}
