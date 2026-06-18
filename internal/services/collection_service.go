package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/services/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
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
	db     *gorm.DB
	engine *workflow.Engine
}

func NewCollectionService(db *gorm.DB, engine *workflow.Engine) CollectionServiceInterface {
	return &collectionService{db: db, engine: engine}
}

func (s *collectionService) syncCollectionWorkflow(ctx context.Context, collectionID uint, status string) {
	if !applyCollectionWorkflow(ctx, s.engine, collectionID, status, nil) {
		utils.Log.WithField("collection_id", collectionID).WithField("status", status).
			Debug("collection workflow sync skipped or failed")
	}
}

func (s *collectionService) getCollectionByID(ctx context.Context, id uint) (*models.Collection, error) {
	var collection models.Collection
	if err := s.db.WithContext(ctx).Preload("WorkflowState").First(&collection, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &collection, nil
}

func (s *collectionService) uniqueSlug(baseSlug string, excludeID uint) string {
	slug := baseSlug
	counter := 1
	for {
		var count int64
		query := s.db.Model(&models.Collection{}).Where("slug = ?", slug)
		if excludeID > 0 {
			query = query.Where("id != ?", excludeID)
		}
		query.Count(&count)
		if count == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
	return slug
}

func (s *collectionService) Create(ctx context.Context, req *dto.CreateCollectionRequest) (*dto.CollectionResponse, error) {
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Title)
	}
	slug = s.uniqueSlug(slug, 0)

	status := req.Status
	if status == "" {
		status = "draft"
	}

	href := req.Href
	if href == "" {
		href = "/shop"
	}

	cta := req.CTALabel
	if cta == "" {
		cta = "Shop collection"
	}

	collection := models.Collection{
		Slug:              slug,
		Eyebrow:           req.Eyebrow,
		Title:             req.Title,
		Description:       req.Description,
		Href:              href,
		ImageURL:          req.ImageURL,
		CTALabel:          cta,
		SortOrder:         req.SortOrder,
		Status:            status,
		PreviewSort:       req.PreviewSort,
		PreviewIsNew:      req.PreviewIsNew,
		PreviewCategoryID: req.PreviewCategoryID,
	}

	if err := s.db.WithContext(ctx).Create(&collection).Error; err != nil {
		return nil, err
	}

	s.syncCollectionWorkflow(ctx, collection.ID, collection.Status)

	loaded, err := s.getCollectionByID(ctx, collection.ID)
	if err != nil {
		return nil, err
	}
	return collectionToResponse(loaded), nil
}

func (s *collectionService) GetByID(ctx context.Context, id uint) (*dto.CollectionResponse, error) {
	collection, err := s.getCollectionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return collectionToResponse(collection), nil
}

func (s *collectionService) List(ctx context.Context, req *dto.ListCollectionsRequest) ([]dto.CollectionResponse, int64, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := s.db.WithContext(ctx).Model(&models.Collection{})

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("title ILIKE ? OR slug ILIKE ? OR eyebrow ILIKE ?", search, search, search)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	var collections []models.Collection
	if err := query.Order("sort_order ASC, created_at DESC").
		Offset(offset).Limit(limit).
		Preload("WorkflowState").
		Find(&collections).Error; err != nil {
		return nil, 0, err
	}

	resp := make([]dto.CollectionResponse, 0, len(collections))
	for i := range collections {
		resp = append(resp, *collectionToResponse(&collections[i]))
	}
	return resp, total, nil
}

func (s *collectionService) Update(ctx context.Context, id uint, req *dto.UpdateCollectionRequest) (*dto.CollectionResponse, error) {
	collection, err := s.getCollectionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Slug != nil && *req.Slug != "" {
		collection.Slug = s.uniqueSlug(*req.Slug, id)
	}
	if req.Eyebrow != nil {
		collection.Eyebrow = *req.Eyebrow
	}
	if req.Title != nil {
		collection.Title = *req.Title
	}
	if req.Description != nil {
		collection.Description = *req.Description
	}
	if req.Href != nil {
		collection.Href = *req.Href
	}
	if req.ImageURL != nil {
		collection.ImageURL = *req.ImageURL
	}
	if req.CTALabel != nil {
		collection.CTALabel = *req.CTALabel
	}
	if req.SortOrder != nil {
		collection.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		collection.Status = *req.Status
	}
	if req.PreviewSort != nil {
		collection.PreviewSort = *req.PreviewSort
	}
	if req.PreviewIsNew != nil {
		collection.PreviewIsNew = req.PreviewIsNew
	}
	if req.PreviewCategoryID != nil {
		collection.PreviewCategoryID = req.PreviewCategoryID
	}

	if err := s.db.WithContext(ctx).Save(collection).Error; err != nil {
		return nil, err
	}

	if req.Status != nil {
		s.syncCollectionWorkflow(ctx, id, *req.Status)
	}

	loaded, err := s.getCollectionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return collectionToResponse(loaded), nil
}

func (s *collectionService) Delete(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&models.Collection{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func collectionToResponse(c *models.Collection) *dto.CollectionResponse {
	resp := &dto.CollectionResponse{
		ID:                c.ID,
		Slug:              c.Slug,
		Eyebrow:           c.Eyebrow,
		Title:             c.Title,
		Description:       c.Description,
		Href:              c.Href,
		ImageURL:          c.ImageURL,
		CTALabel:          c.CTALabel,
		SortOrder:         c.SortOrder,
		Status:            c.Status,
		PreviewSort:       c.PreviewSort,
		PreviewIsNew:      c.PreviewIsNew,
		PreviewCategoryID: c.PreviewCategoryID,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
	if c.WorkflowState != nil {
		resp.WorkflowState = dto.ToStateView(c.WorkflowState)
	}
	return resp
}
