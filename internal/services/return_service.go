package services

import (
	"context"

	appreturn "github.com/alireza-akbarzadeh/luxe/internal/application/returnorder"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type ReturnServiceInterface interface {
	Create(ctx context.Context, userID uint, req dto.CreateReturnRequest) (*models.Return, error)
	GetByID(ctx context.Context, returnID, userID uint, isAdmin bool) (*models.Return, error)
	ListForUser(ctx context.Context, userID uint, limit, offset int) ([]models.Return, int64, error)
	ListAdmin(ctx context.Context, filters dto.AdminReturnListFilters) ([]models.Return, int64, error)
	PerformTransition(ctx context.Context, returnID uint, event, note, actorRole string, actorID *uint) (*workflow.TransitionResult, error)
}

type returnService struct {
	commands *appreturn.Commands
	queries  *appreturn.Queries
}

func NewReturnService(db *gorm.DB, engine *workflow.Engine) ReturnServiceInterface {
	repo := postgres.NewReturnRepository(db)
	return &returnService{
		commands: appreturn.NewCommands(repo, engine),
		queries:  appreturn.NewQueries(repo),
	}
}

func (s *returnService) Create(ctx context.Context, userID uint, req dto.CreateReturnRequest) (*models.Return, error) {
	return s.commands.Create(ctx, userID, req)
}

func (s *returnService) GetByID(ctx context.Context, returnID, userID uint, isAdmin bool) (*models.Return, error) {
	return s.queries.GetByID(ctx, returnID, userID, isAdmin)
}

func (s *returnService) ListForUser(ctx context.Context, userID uint, limit, offset int) ([]models.Return, int64, error) {
	return s.queries.ListForUser(ctx, userID, limit, offset)
}

func (s *returnService) ListAdmin(ctx context.Context, filters dto.AdminReturnListFilters) ([]models.Return, int64, error) {
	return s.queries.ListAdmin(ctx, filters)
}

func (s *returnService) PerformTransition(ctx context.Context, returnID uint, event, note, actorRole string, actorID *uint) (*workflow.TransitionResult, error) {
	return s.commands.PerformTransition(ctx, returnID, event, note, actorRole, actorID)
}
