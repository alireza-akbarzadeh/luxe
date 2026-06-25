package services

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	appcollection "github.com/alireza-akbarzadeh/luxe/internal/application/collection"
	domaincollection "github.com/alireza-akbarzadeh/luxe/internal/domain/collection"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

type CollectionServiceInterface interface {
	Create(ctx context.Context, req *dto.CreateCollectionRequest) (*dto.CollectionResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.CollectionResponse, error)
	List(ctx context.Context, req *dto.ListCollectionsRequest) ([]dto.CollectionResponse, int64, error)
	Update(ctx context.Context, id uint, req *dto.UpdateCollectionRequest) (*dto.CollectionResponse, error)
	Delete(ctx context.Context, id uint) error
}

type collectionService struct {
	engine   *workflow.Engine
	commands *appcollection.Commands
	queries  *appcollection.Queries
}

func NewCollectionService(db *gorm.DB, engine *workflow.Engine) CollectionServiceInterface {
	repo := postgres.NewCollectionRepository(db)
	return &collectionService{
		engine:   engine,
		commands: appcollection.NewCommands(domaincollection.NewService(), repo),
		queries:  appcollection.NewQueries(repo),
	}
}

func (s *collectionService) syncCollectionWorkflow(ctx context.Context, collectionID uint, status string) {
	if !applyCollectionWorkflow(ctx, s.engine, collectionID, status, nil) {
		utils.Log.WithField("collection_id", collectionID).WithField("status", status).
			Debug("collection workflow sync skipped or failed")
	}
}

func (s *collectionService) Create(ctx context.Context, req *dto.CreateCollectionRequest) (*dto.CollectionResponse, error) {
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Title)
	}
	unique, err := s.commands.UniqueSlug(ctx, slug, 0)
	if err != nil {
		return nil, err
	}

	collection, err := s.commands.Create(ctx, req, unique)
	if err != nil {
		if errors.Is(err, domaincollection.ErrInvalidTitle) {
			return nil, err
		}
		return nil, err
	}

	s.syncCollectionWorkflow(ctx, collection.ID, collection.Status)

	loaded, err := s.queries.GetByID(ctx, collection.ID)
	if err != nil {
		return nil, err
	}
	return appcollection.ToResponse(loaded), nil
}

func (s *collectionService) GetByID(ctx context.Context, id uint) (*dto.CollectionResponse, error) {
	collection, err := s.queries.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return appcollection.ToResponse(collection), nil
}

func (s *collectionService) List(ctx context.Context, req *dto.ListCollectionsRequest) ([]dto.CollectionResponse, int64, error) {
	collections, total, _, _, err := s.queries.List(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.CollectionResponse, 0, len(collections))
	for i := range collections {
		resp = append(resp, *appcollection.ToResponse(&collections[i]))
	}
	return resp, total, nil
}

func (s *collectionService) Update(ctx context.Context, id uint, req *dto.UpdateCollectionRequest) (*dto.CollectionResponse, error) {
	collection, err := s.queries.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	newSlug := ""
	if req.Slug != nil && *req.Slug != "" {
		unique, err := s.commands.UniqueSlug(ctx, *req.Slug, id)
		if err != nil {
			return nil, err
		}
		newSlug = unique
	}

	appcollection.ApplyUpdateDTO(collection, req, newSlug)

	if err := s.commands.Update(ctx, collection); err != nil {
		return nil, err
	}

	if req.Status != nil {
		s.syncCollectionWorkflow(ctx, id, *req.Status)
	}

	loaded, err := s.queries.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return appcollection.ToResponse(loaded), nil
}

func (s *collectionService) Delete(ctx context.Context, id uint) error {
	rows, err := s.commands.Delete(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
