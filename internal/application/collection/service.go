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
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
	"time"
)

// Service handles collection HTTP-oriented use cases (DTO mapping and workflow sync).
type Service struct {
	engine   *workflow.Engine
	commands *Commands
	queries  *Queries
	products *postgres.ProductRepository
}

// NewService wires collection application use cases.
func NewService(db *gorm.DB, engine *workflow.Engine) *Service {
	repo := postgres.NewCollectionRepository(db)
	return &Service{
		engine:   engine,
		commands: NewCommands(domaincollection.NewService(), repo),
		queries:  NewQueries(repo),
		products: postgres.NewProductRepository(db),
	}
}

func (s *Service) syncCollectionWorkflow(ctx context.Context, collectionID uint, status string) {
	if !appworkflow.ApplyCollectionWorkflow(ctx, s.engine, collectionID, status, nil) {
		utils.Log.WithField("collection_id", collectionID).WithField("status", status).
			Debug("collection workflow sync skipped or failed")
	}
}

func isCollectionLive(collection *models.Collection) bool {
	if collection == nil || collection.Status != "active" {
		return false
	}
	now := time.Now().UTC()
	if collection.StartsAt != nil && collection.StartsAt.After(now) {
		return false
	}
	if collection.EndsAt != nil && collection.EndsAt.Before(now) {
		return false
	}
	return true
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

// GetBySlug returns a live collection API response by current or redirected slug.
func (s *Service) GetBySlug(ctx context.Context, slug string) (*dto.CollectionResponse, error) {
	collection, err := s.queries.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("collection not found")
		}
		return nil, err
	}
	if !isCollectionLive(collection) {
		return nil, utils.ErrNotFound("collection not found")
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

	oldSlug := collection.Slug
	if err := ApplyUpdateDTO(collection, req, newSlug); err != nil {
		return nil, err
	}

	if err := s.commands.Update(ctx, collection, req, oldSlug); err != nil {
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

// ListResolvedProducts resolves the products for a collection slug.
func (s *Service) ListResolvedProducts(
	ctx context.Context,
	slug string,
	req *dto.CollectionProductsRequest,
) (*dto.ProductListData, *dto.CollectionResponse, error) {
	collection, err := s.queries.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, utils.ErrNotFound("collection not found")
		}
		return nil, nil, err
	}
	if !isCollectionLive(collection) {
		return nil, nil, utils.ErrNotFound("collection not found")
	}
	data, err := s.resolveCollectionProducts(ctx, collection, req)
	if err != nil {
		return nil, nil, err
	}
	return data, ToResponse(collection), nil
}

// PreviewProducts resolves products for a saved collection.
func (s *Service) PreviewProducts(
	ctx context.Context,
	id uint,
	req *dto.CollectionProductsRequest,
) (*dto.ProductListData, error) {
	collection, err := s.queries.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("collection not found")
		}
		return nil, err
	}
	return s.resolveCollectionProducts(ctx, collection, req)
}

// ValidateRules previews a transient collection rule set without saving it.
func (s *Service) ValidateRules(
	ctx context.Context,
	id uint,
	req *dto.CollectionRulesValidationRequest,
) (*dto.ProductListData, error) {
	collection, err := s.queries.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("collection not found")
		}
		return nil, err
	}
	return s.previewRulesOnCollection(ctx, collection, req)
}

// ValidateRulesTransient previews rules without a saved collection (create flow).
func (s *Service) ValidateRulesTransient(
	ctx context.Context,
	req *dto.CollectionRulesValidationRequest,
) (*dto.ProductListData, error) {
	mode := req.Mode
	if mode == "" {
		mode = "dynamic"
	}
	collection := &models.Collection{
		Mode:           mode,
		CollectionType: legacyCollectionType(mode, ""),
		Status:         "draft",
	}
	return s.previewRulesOnCollection(ctx, collection, req)
}

func (s *Service) previewRulesOnCollection(
	ctx context.Context,
	collection *models.Collection,
	req *dto.CollectionRulesValidationRequest,
) (*dto.ProductListData, error) {
	if req.Mode != "" {
		collection.Mode = req.Mode
		collection.CollectionType = legacyCollectionType(req.Mode, collection.CollectionType)
	}
	rules := req.Rules
	if rules == nil {
		rules = unmarshalRules(collection.RulesJSON)
	}
	if err := ValidateCollectionRules(collection.Mode, rules); err != nil {
		return nil, err
	}
	if req.Rules != nil {
		rulesJSON, err := marshalRules(req.Rules)
		if err != nil {
			return nil, err
		}
		collection.RulesJSON = rulesJSON
	}
	if len(req.Overrides) > 0 {
		collection.Products = make([]models.CollectionProduct, 0, len(req.Overrides))
		for _, override := range req.Overrides {
			collection.Products = append(collection.Products, models.CollectionProduct{
				CollectionID: collection.ID,
				ProductID:    override.ProductID,
				Position:     override.Position,
				SortOrder:    override.Position,
				IsPinned:     override.IsPinned,
				IsHidden:     override.IsHidden,
				BoostScore:   override.BoostScore,
			})
		}
	}
	previewLimit := 4
	if req.Limit > 0 {
		previewLimit = req.Limit
	}
	return s.resolveCollectionProducts(ctx, collection, &dto.CollectionProductsRequest{
		Limit:  previewLimit,
		Offset: 0,
	})
}

// ActivateDueCollections promotes scheduled collections whose starts_at has passed.
func (s *Service) ActivateDueCollections(ctx context.Context) (int64, error) {
	n, err := s.commands.ActivateDue(ctx)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// ExpireEndedCollections deactivates active collections past ends_at.
func (s *Service) ExpireEndedCollections(ctx context.Context) (int64, error) {
	return s.commands.ExpireEnded(ctx)
}
