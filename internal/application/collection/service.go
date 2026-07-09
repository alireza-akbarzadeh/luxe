package collection

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	appcategory "github.com/alireza-akbarzadeh/luxe/internal/application/category"
	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	domaincollection "github.com/alireza-akbarzadeh/luxe/internal/domain/collection"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service handles collection HTTP-oriented use cases (DTO mapping and workflow sync).
type Service struct {
	engine   *workflow.Engine
	commands *Commands
	queries  *Queries
}

// NewService wires collection application use cases.
func NewService(db *gorm.DB, engine *workflow.Engine) *Service {
	repo := postgres.NewCollectionRepository(db)
	return &Service{
		engine:   engine,
		commands: NewCommands(domaincollection.NewService(), repo),
		queries:  NewQueries(repo),
	}
}

func (s *Service) syncCollectionWorkflow(ctx context.Context, collectionID uint, status string) {
	if !appworkflow.ApplyCollectionWorkflow(ctx, s.engine, collectionID, status, nil) {
		utils.Log.WithField("collection_id", collectionID).WithField("status", status).
			Debug("collection workflow sync skipped or failed")
	}
}

// Create creates a collection and returns an API response DTO.
func (s *Service) Create(ctx context.Context, req *dto.CreateCollectionRequest) (*dto.CollectionResponse, error) {
	slug := req.Slug
	if slug == "" {
		slug = appcategory.GenerateSlug(req.Title)
	}
	unique, err := s.commands.UniqueSlug(ctx, slug, 0)
	if err != nil {
		return nil, err
	}

	collection, err := s.commands.Create(ctx, req, unique)
	if err != nil {
		return nil, err
	}

	s.syncCollectionWorkflow(ctx, collection.ID, collection.Status)

	loaded, err := s.queries.GetByID(ctx, collection.ID)
	if err != nil {
		return nil, err
	}
	return ToResponse(loaded), nil
}

// GetByID returns a collection API response by ID.
func (s *Service) GetByID(ctx context.Context, id uint) (*dto.CollectionResponse, error) {
	collection, err := s.queries.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("collection not found")
		}
		return nil, err
	}
	return ToResponse(collection), nil
}

// List returns paginated collection API responses.
func (s *Service) List(ctx context.Context, req *dto.ListCollectionsRequest) ([]dto.CollectionResponse, int64, error) {
	collections, total, _, _, err := s.queries.List(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.CollectionResponse, 0, len(collections))
	for i := range collections {
		resp = append(resp, *ToResponse(&collections[i]))
	}
	return resp, total, nil
}

// Update updates a collection and returns an API response DTO.
func (s *Service) Update(ctx context.Context, id uint, req *dto.UpdateCollectionRequest) (*dto.CollectionResponse, error) {
	collection, err := s.queries.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("collection not found")
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

	if err := ApplyUpdateDTO(collection, req, newSlug); err != nil {
		return nil, err
	}

	if err := s.commands.Update(ctx, collection, req); err != nil {
		return nil, err
	}

	if req.Status != nil {
		s.syncCollectionWorkflow(ctx, id, *req.Status)
	}

	loaded, err := s.queries.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ToResponse(loaded), nil
}

// Delete removes a collection by ID.
func (s *Service) Delete(ctx context.Context, id uint) error {
	rows, err := s.commands.Delete(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return utils.ErrNotFound("collection not found")
	}
	return nil
}
