package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	appcatalog "github.com/alireza-akbarzadeh/luxe/internal/application/catalog"
	domaincatalog "github.com/alireza-akbarzadeh/luxe/internal/domain/catalog"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// setProductState syncs a product status into the workflow engine (best-effort).
func (s *productService) setProductState(ctx context.Context, productID uint, status, actorRole string, actorID *uint) {
	if !applyProductWorkflow(ctx, s.engine, productID, status, actorRole, actorID) {
		utils.Log.WithField("product_id", productID).WithField("status", status).
			Debug("product workflow sync skipped or failed")
	}
}

type ProductServiceInterface interface {
	List(limit, offset int, filters dto.ProductListFilters) ([]*models.Product, int64, error)
	BulkCreate(products []dto.CreateProductRequest) ([]*models.Product, error)
	GetByID(id uint) (*models.Product, error)
	GetBySlug(slug string) (*models.Product, error)
	Create(req dto.CreateProductRequest) (*models.Product, error)
	Update(productID uint, req dto.UpdateProductRequest) (*models.Product, error)
	Delete(id uint) error
	BulkDelete(productIDs []uint) error
	CheckLowStockAndAlert() error
	GetRelated(productID uint, limit int) ([]*models.Product, error)
	GetSuggestions(productIDs []uint, limit int) ([]*models.Product, error)
	GetByStoreID(storeID uint, limit, offset int, filters dto.ProductListFilters) ([]*models.Product, int64, error)
	AvailableTransitions(ctx context.Context, productID uint) (*models.WorkflowState, []models.WorkflowTransition, error)
	PerformTransition(ctx context.Context, productID uint, event, note, actorRole string, actorID *uint) (*workflow.TransitionResult, error)
	SetInventory(inventory InventoryServiceInterface)
}

type productService struct {
	engine    *workflow.Engine
	inventory InventoryServiceInterface
	catalog   *appcatalog.Commands
	queries   *appcatalog.Queries
}

func NewProductService(db *gorm.DB, engine *workflow.Engine) ProductServiceInterface {
	repo := postgres.NewProductRepository(db)
	return &productService{
		engine:  engine,
		catalog: appcatalog.NewCommands(domaincatalog.NewService(), repo, repo),
		queries: appcatalog.NewQueries(repo, repo),
	}
}

// SetInventory wires the inventory ledger after DI construction.
func (s *productService) SetInventory(inventory InventoryServiceInterface) {
	s.inventory = inventory
}

// UniqSlug checks and modifies slug to be unique.
func (s *productService) UniqSlug(baseSlug string, excludeID uint) string {
	slug := baseSlug
	counter := 1
	ctx := context.Background()
	for {
		taken, err := s.catalog.SlugTaken(ctx, slug, excludeID)
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

func (s *productService) Create(req dto.CreateProductRequest) (*models.Product, error) {
	ctx := context.Background()
	slug := s.UniqSlug(generateSlug(req.Name), 0)

	product, err := s.catalog.PrepareCreate(req, slug)
	if err != nil {
		return nil, utils.ErrValidationFailed(err.Error())
	}
	product.SearchDocument = s.queries.BuildSearchDocument(ctx, product)

	if err := s.catalog.PersistCreate(ctx, product); err != nil {
		return nil, utils.ErrInternal(err)
	}
	s.setProductState(ctx, product.ID, product.Status, constants.RoleAdmin, nil)
	if s.inventory != nil {
		_ = s.inventory.RecordInitialStock(ctx, product.ID, product.Stock)
	}
	return product, nil
}

func (s *productService) GetByID(id uint) (*models.Product, error) {
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

func (s *productService) GetBySlug(slug string) (*models.Product, error) {
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

func (s *productService) Update(id uint, req dto.UpdateProductRequest) (*models.Product, error) {
	ctx := context.Background()
	product, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	newSlug := ""
	if req.Name != nil {
		newSlug = s.UniqSlug(generateSlug(*req.Name), id)
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

	appcatalog.ApplyUpdateDTO(product, req, newSlug)
	product.SearchDocument = s.queries.BuildSearchDocument(ctx, product)

	stockUpdate := req.Stock

	if err := s.catalog.Save(ctx, product); err != nil {
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
		} else if err := s.catalog.UpdateStockColumn(ctx, id, *stockUpdate); err != nil {
			return nil, utils.ErrInternal(err)
		}
	}

	if req.Attributes != nil {
		if err := s.catalog.ReplaceAttributes(ctx, id, *req.Attributes); err != nil {
			return nil, utils.ErrInternal(err)
		}
	}

	return s.GetByID(id)
}

func (s *productService) Delete(id uint) error {
	rows, err := s.catalog.Delete(context.Background(), id)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("product not found")
	}
	return nil
}

func (s *productService) List(limit, offset int, filters dto.ProductListFilters) ([]*models.Product, int64, error) {
	return s.queries.ListDetailed(context.Background(), limit, offset, filters)
}

func (s *productService) BulkCreate(products []dto.CreateProductRequest) ([]*models.Product, error) {
	if len(products) == 0 {
		return nil, utils.ErrBadRequest("no products provided")
	}

	toCreate := make([]*models.Product, 0, len(products))
	for _, p := range products {
		slug := s.UniqSlug(generateSlug(p.Name), 0)
		toCreate = append(toCreate, appcatalog.BuildBulkCreateModel(p, slug))
	}

	ctx := context.Background()
	if err := s.catalog.BulkCreate(ctx, toCreate); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return toCreate, nil
}

func (s *productService) BulkDelete(productIDs []uint) error {
	if len(productIDs) == 0 {
		return utils.ErrBadRequest("no product IDs provided")
	}
	rows, err := s.catalog.BulkDelete(context.Background(), productIDs)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("products not found")
	}
	return nil
}

func (s *productService) CheckLowStockAndAlert() error {
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

func (s *productService) GetRelated(productID uint, limit int) ([]*models.Product, error) {
	return s.queries.GetRelated(context.Background(), productID, limit)
}

func (s *productService) GetSuggestions(productIDs []uint, limit int) ([]*models.Product, error) {
	return s.queries.GetSuggestions(context.Background(), productIDs, limit)
}

func (s *productService) GetByStoreID(storeID uint, limit, offset int, filters dto.ProductListFilters) ([]*models.Product, int64, error) {
	filters.StoreID = &storeID
	return s.List(limit, offset, filters)
}

func (s *productService) ensureProductExists(ctx context.Context, productID uint) error {
	exists, err := s.queries.ExistsByID(ctx, productID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if !exists {
		return utils.ErrNotFound("product not found")
	}
	return nil
}

func (s *productService) AvailableTransitions(ctx context.Context, productID uint) (*models.WorkflowState, []models.WorkflowTransition, error) {
	if s.engine == nil {
		return nil, nil, utils.ErrInternal(errors.New("workflow engine not configured"))
	}
	if err := s.ensureProductExists(ctx, productID); err != nil {
		return nil, nil, err
	}
	return s.engine.AvailableTransitions(ctx, constants.WorkflowEntityProduct, productID)
}

func (s *productService) PerformTransition(
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
