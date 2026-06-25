package catalog

import (
	"context"
	"errors"
	"fmt"

	appcategory "github.com/alireza-akbarzadeh/luxe/internal/application/category"
	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	domaincatalog "github.com/alireza-akbarzadeh/luxe/internal/domain/catalog"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// InventoryHook records stock changes from product create/update flows.
type InventoryHook interface {
	RecordInitialStock(ctx context.Context, productID uint, quantity int) error
	SetAbsoluteStock(ctx context.Context, productID uint, newStock int, actorUserID *uint, note, adjustmentType string) error
}

// Service orchestrates product catalog use cases for HTTP handlers.
type Service struct {
	engine    *workflow.Engine
	inventory InventoryHook
	commands  *Commands
	queries   *Queries
}

// NewService wires product catalog commands and queries.
func NewService(db *gorm.DB, engine *workflow.Engine) *Service {
	repo := postgres.NewProductRepository(db)
	return &Service{
		engine:   engine,
		commands: NewCommands(domaincatalog.NewService(), repo, repo),
		queries:  NewQueries(repo, repo),
	}
}

// SetInventory wires the inventory ledger after DI construction.
func (s *Service) SetInventory(inventory InventoryHook) {
	s.inventory = inventory
}

func (s *Service) setProductState(ctx context.Context, productID uint, status, actorRole string, actorID *uint) {
	if !appworkflow.ApplyProductWorkflow(ctx, s.engine, productID, status, actorRole, actorID) {
		utils.Log.WithField("product_id", productID).WithField("status", status).
			Debug("product workflow sync skipped or failed")
	}
}

// UniqSlug checks and modifies slug to be unique.
func (s *Service) UniqSlug(baseSlug string, excludeID uint) string {
	slug := baseSlug
	counter := 1
	ctx := context.Background()
	for {
		taken, err := s.commands.SlugTaken(ctx, slug, excludeID)
		if err != nil {
			utils.Log.WithError(err).Warn("catalog slug check failed")
			break
		}
		if !taken {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
	return slug
}

func (s *Service) Create(req dto.CreateProductRequest) (*models.Product, error) {
	ctx := context.Background()
	slug := s.UniqSlug(appcategory.GenerateSlug(req.Name), 0)

	product, err := s.commands.PrepareCreate(req, slug)
	if err != nil {
		return nil, utils.ErrValidationFailed(err.Error())
	}
	product.SearchDocument = s.queries.BuildSearchDocument(ctx, product)

	if err := s.commands.PersistCreate(ctx, product); err != nil {
		return nil, utils.ErrInternal(err)
	}
	s.setProductState(ctx, product.ID, product.Status, constants.RoleAdmin, nil)
	if s.inventory != nil {
		_ = s.inventory.RecordInitialStock(ctx, product.ID, product.Stock)
	}
	return product, nil
}

func (s *Service) GetByID(id uint) (*models.Product, error) {
	ctx := context.Background()
	product, err := s.queries.GetDetailedByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return product, nil
}

func (s *Service) GetBySlug(slug string) (*models.Product, error) {
	ctx := context.Background()
	product, err := s.queries.GetDetailedBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return product, nil
}

func (s *Service) Update(id uint, req dto.UpdateProductRequest) (*models.Product, error) {
	ctx := context.Background()
	product, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	newSlug := ""
	if req.Name != nil {
		newSlug = s.UniqSlug(appcategory.GenerateSlug(*req.Name), id)
	}
	if req.SKU != nil {
		taken, err := s.queries.SKUTaken(ctx, *req.SKU, id)
		if err != nil {
			return nil, utils.ErrInternal(err)
		}
		if taken {
			return nil, utils.ErrConflict("SKU already exists")
		}
	}

	ApplyUpdateDTO(product, req, newSlug)
	product.SearchDocument = s.queries.BuildSearchDocument(ctx, product)

	stockUpdate := req.Stock

	if err := s.commands.Save(ctx, product); err != nil {
		return nil, utils.ErrInternal(err)
	}
	if req.Status != nil {
		s.setProductState(ctx, product.ID, product.Status, constants.RoleAdmin, nil)
	}

	if stockUpdate != nil {
		if s.inventory != nil && product.TrackInventory {
			if err := s.inventory.SetAbsoluteStock(ctx, id, *stockUpdate, nil, "product update", constants.InventoryAdjAdminSet); err != nil {
				return nil, err
			}
		} else if err := s.commands.UpdateStockColumn(ctx, id, *stockUpdate); err != nil {
			return nil, utils.ErrInternal(err)
		}
	}

	if req.Attributes != nil {
		if err := s.commands.ReplaceAttributes(ctx, id, *req.Attributes); err != nil {
			return nil, utils.ErrInternal(err)
		}
	}

	return s.GetByID(id)
}

func (s *Service) Delete(id uint) error {
	rows, err := s.commands.Delete(context.Background(), id)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("product not found")
	}
	return nil
}

func (s *Service) List(limit, offset int, filters dto.ProductListFilters) ([]*models.Product, int64, error) {
	return s.queries.ListDetailed(context.Background(), limit, offset, filters)
}

func (s *Service) BulkCreate(products []dto.CreateProductRequest) ([]*models.Product, error) {
	if len(products) == 0 {
		return nil, utils.ErrBadRequest("no products provided")
	}

	toCreate := make([]*models.Product, 0, len(products))
	for _, p := range products {
		slug := s.UniqSlug(appcategory.GenerateSlug(p.Name), 0)
		toCreate = append(toCreate, BuildBulkCreateModel(p, slug))
	}

	ctx := context.Background()
	if err := s.commands.BulkCreate(ctx, toCreate); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return toCreate, nil
}

func (s *Service) BulkDelete(productIDs []uint) error {
	if len(productIDs) == 0 {
		return utils.ErrBadRequest("no product IDs provided")
	}
	rows, err := s.commands.BulkDelete(context.Background(), productIDs)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("products not found")
	}
	return nil
}

func (s *Service) CheckLowStockAndAlert() error {
	products, err := s.queries.FindLowStockActive(context.Background())
	if err != nil {
		return utils.ErrInternal(err)
	}

	if len(products) == 0 {
		utils.Log.Info("Low stock check: no products below threshold")
		return nil
	}

	for _, p := range products {
		utils.Log.Warnf("LOW STOCK ALERT: Product ID=%d, Name=%s, Stock=%d, Threshold=%d",
			p.ID, p.Name, p.Stock, p.LowStockThreshold)
	}

	return nil
}

func (s *Service) GetRelated(productID uint, limit int) ([]*models.Product, error) {
	return s.queries.GetRelated(context.Background(), productID, limit)
}

func (s *Service) GetSuggestions(productIDs []uint, limit int) ([]*models.Product, error) {
	return s.queries.GetSuggestions(context.Background(), productIDs, limit)
}

func (s *Service) GetByStoreID(storeID uint, limit, offset int, filters dto.ProductListFilters) ([]*models.Product, int64, error) {
	filters.StoreID = &storeID
	return s.List(limit, offset, filters)
}

func (s *Service) ensureProductExists(ctx context.Context, productID uint) error {
	exists, err := s.queries.ExistsByID(ctx, productID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if !exists {
		return utils.ErrNotFound("product not found")
	}
	return nil
}

func (s *Service) AvailableTransitions(ctx context.Context, productID uint) (*models.WorkflowState, []models.WorkflowTransition, error) {
	if s.engine == nil {
		return nil, nil, utils.ErrInternal(errors.New("workflow engine not configured"))
	}
	if err := s.ensureProductExists(ctx, productID); err != nil {
		return nil, nil, err
	}
	return s.engine.AvailableTransitions(ctx, constants.WorkflowEntityProduct, productID)
}

func (s *Service) PerformTransition(
	ctx context.Context,
	productID uint,
	event, note, actorRole string,
	actorID *uint,
) (*workflow.TransitionResult, error) {
	if s.engine == nil {
		return nil, utils.ErrInternal(errors.New("workflow engine not configured"))
	}
	if err := s.ensureProductExists(ctx, productID); err != nil {
		return nil, err
	}
	return s.engine.Transition(ctx, workflow.TransitionRequest{
		WorkflowKey: constants.WorkflowEntityProduct,
		EntityID:    productID,
		Event:       event,
		ActorID:     actorID,
		ActorRole:   actorRole,
		Note:        note,
	})
}
