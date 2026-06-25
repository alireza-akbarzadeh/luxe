package services

import (
	"context"

	appmenu "github.com/alireza-akbarzadeh/luxe/internal/application/menu"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type UserMenuServicesInterface interface {
	GetAllGroups() ([]models.MenuGroup, error)
	GetGroupByID(id uint) (*models.MenuGroup, error)
	CreateGroup(req *dto.CreateMenuGroupRequest) (*models.MenuGroup, error)
	UpdateGroup(id uint, req *dto.UpdateMenuGroupRequest) (*models.MenuGroup, error)
	DeleteGroup(id uint) error

	GetAllItems(flat bool) ([]models.MenuItem, error)
	GetItemByID(id uint) (*models.MenuItem, error)
	CreateItem(req *dto.CreateMenuItemRequest) (*models.MenuItem, error)
	UpdateItem(id uint, req *dto.UpdateMenuItemRequest) (*models.MenuItem, error)
	DeleteItem(id uint) error

	GetUserMenu(ctx context.Context, userRole string, search string) ([]dto.SidebarGroup, error)
	GetUserMenuStructure(ctx context.Context, userRole string, search string) ([]dto.MenuGroupResponse, error)
}

type userMenuService struct {
	commands *appmenu.Commands
	queries  *appmenu.Queries
}

func NewMenuService(db *gorm.DB) UserMenuServicesInterface {
	repo := postgres.NewMenuRepository(db)
	return &userMenuService{
		commands: appmenu.NewCommands(repo),
		queries:  appmenu.NewQueries(repo),
	}
}

func (s *userMenuService) GetAllGroups() ([]models.MenuGroup, error) {
	return s.queries.ListGroups(context.Background())
}

func (s *userMenuService) GetGroupByID(id uint) (*models.MenuGroup, error) {
	return s.queries.GetGroupByID(context.Background(), id)
}

func (s *userMenuService) CreateGroup(req *dto.CreateMenuGroupRequest) (*models.MenuGroup, error) {
	return s.commands.CreateGroup(context.Background(), req)
}

func (s *userMenuService) UpdateGroup(id uint, req *dto.UpdateMenuGroupRequest) (*models.MenuGroup, error) {
	return s.commands.UpdateGroup(context.Background(), id, req)
}

func (s *userMenuService) DeleteGroup(id uint) error {
	return s.commands.DeleteGroup(context.Background(), id)
}

func (s *userMenuService) GetAllItems(flat bool) ([]models.MenuItem, error) {
	return s.queries.ListAllItems(context.Background(), flat)
}

func (s *userMenuService) GetItemByID(id uint) (*models.MenuItem, error) {
	return s.queries.GetItemByID(context.Background(), id)
}

func (s *userMenuService) CreateItem(req *dto.CreateMenuItemRequest) (*models.MenuItem, error) {
	return s.commands.CreateItem(context.Background(), req)
}

func (s *userMenuService) UpdateItem(id uint, req *dto.UpdateMenuItemRequest) (*models.MenuItem, error) {
	return s.commands.UpdateItem(context.Background(), id, req)
}

func (s *userMenuService) DeleteItem(id uint) error {
	return s.commands.DeleteItem(context.Background(), id)
}

func (s *userMenuService) GetUserMenu(ctx context.Context, userRole string, search string) ([]dto.SidebarGroup, error) {
	return s.queries.GetUserMenu(ctx, userRole, search)
}

func (s *userMenuService) GetUserMenuStructure(ctx context.Context, userRole string, search string) ([]dto.MenuGroupResponse, error) {
	return s.queries.GetUserMenuStructure(ctx, userRole, search)
}
