package services

import (
	"context"

	appnavmenu "github.com/alireza-akbarzadeh/luxe/internal/application/navmenu"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"gorm.io/gorm"
)

type NavMenuServiceInterface interface {
	GetAll(ctx context.Context) ([]dto.NavItemResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.NavItemResponse, error)
	Create(ctx context.Context, req *dto.UpsertNavMenuRequest) (*dto.NavItemResponse, error)
	Update(ctx context.Context, id uint, req *dto.UpsertNavMenuRequest) (*dto.NavItemResponse, error)
	Reorder(ctx context.Context, req *dto.ReorderNavMenusRequest) error
	Delete(ctx context.Context, id uint) error
}

type navMenuService struct {
	commands *appnavmenu.Commands
	queries  *appnavmenu.Queries
}

func NewNavMenuService(db *gorm.DB) NavMenuServiceInterface {
	repo := postgres.NewNavMenuRepository(db)
	return &navMenuService{
		commands: appnavmenu.NewCommands(repo),
		queries:  appnavmenu.NewQueries(repo),
	}
}

func (s *navMenuService) GetAll(ctx context.Context) ([]dto.NavItemResponse, error) {
	return s.queries.GetAll(ctx)
}

func (s *navMenuService) GetByID(ctx context.Context, id uint) (*dto.NavItemResponse, error) {
	return s.queries.GetByID(ctx, id)
}

func (s *navMenuService) Create(ctx context.Context, req *dto.UpsertNavMenuRequest) (*dto.NavItemResponse, error) {
	return s.commands.Create(ctx, req)
}

func (s *navMenuService) Update(ctx context.Context, id uint, req *dto.UpsertNavMenuRequest) (*dto.NavItemResponse, error) {
	return s.commands.Update(ctx, id, req)
}

func (s *navMenuService) Reorder(ctx context.Context, req *dto.ReorderNavMenusRequest) error {
	return s.commands.Reorder(ctx, req)
}

func (s *navMenuService) Delete(ctx context.Context, id uint) error {
	return s.commands.Delete(ctx, id)
}
