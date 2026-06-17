package services

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/services/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

// ReturnServiceInterface manages customer return/refund requests via the return workflow.
type ReturnServiceInterface interface {
	Create(ctx context.Context, userID uint, req dto.CreateReturnRequest) (*models.Return, error)
	GetByID(ctx context.Context, returnID, userID uint, isAdmin bool) (*models.Return, error)
	ListForUser(ctx context.Context, userID uint, limit, offset int) ([]models.Return, int64, error)
	ListAdmin(ctx context.Context, filters dto.AdminReturnListFilters) ([]models.Return, int64, error)
	PerformTransition(ctx context.Context, returnID uint, event, note, actorRole string, actorID *uint) (*workflow.TransitionResult, error)
}

type returnService struct {
	db     *gorm.DB
	engine *workflow.Engine
}

func NewReturnService(db *gorm.DB, engine *workflow.Engine) ReturnServiceInterface {
	return &returnService{db: db, engine: engine}
}

var returnEligibleOrderStatuses = map[string]bool{
	constants.OrderStatusDelivered: true,
	"completed":                    true,
}

func (s *returnService) Create(ctx context.Context, userID uint, req dto.CreateReturnRequest) (*models.Return, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", req.OrderID, userID).
		First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("order not found")
		}
		return nil, utils.ErrInternal(err)
	}

	if !returnEligibleOrderStatuses[order.Status] {
		return nil, utils.ErrBadRequest("returns are only allowed for delivered or completed orders")
	}

	var existing int64
	if err := s.db.WithContext(ctx).Model(&models.Return{}).
		Where("order_id = ? AND user_id = ? AND status NOT IN ?", req.OrderID, userID,
			[]string{"rejected", "closed", "refunded"}).
		Count(&existing).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	if existing > 0 {
		return nil, utils.ErrConflict("an open return already exists for this order")
	}

	ret := &models.Return{
		OrderID:      req.OrderID,
		UserID:       userID,
		Reason:       req.Reason,
		Status:       "requested",
		RefundAmount: order.TotalAmount,
	}

	if err := s.db.WithContext(ctx).Create(ret).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	syncWorkflowState(ctx, s.engine, constants.WorkflowEntityReturn, ret.ID, "requested", "return_requested", &userID)
	return ret, nil
}

func (s *returnService) GetByID(ctx context.Context, returnID, userID uint, isAdmin bool) (*models.Return, error) {
	q := s.db.WithContext(ctx).Preload("Order").Preload("WorkflowState").Where("id = ?", returnID)
	if !isAdmin {
		q = q.Where("user_id = ?", userID)
	}

	var ret models.Return
	if err := q.First(&ret).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("return not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &ret, nil
}

func (s *returnService) ListForUser(ctx context.Context, userID uint, limit, offset int) ([]models.Return, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	q := s.db.WithContext(ctx).Model(&models.Return{}).Where("user_id = ?", userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	var returns []models.Return
	if err := q.Preload("Order").Preload("WorkflowState").
		Order("created_at DESC").Limit(limit).Offset(offset).
		Find(&returns).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return returns, total, nil
}

func (s *returnService) ListAdmin(ctx context.Context, filters dto.AdminReturnListFilters) ([]models.Return, int64, error) {
	limit, offset := filters.Limit, filters.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	q := s.db.WithContext(ctx).Model(&models.Return{})
	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	}
	if filters.UserID != nil {
		q = q.Where("user_id = ?", *filters.UserID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	var returns []models.Return
	if err := q.Preload("Order").Preload("WorkflowState").
		Order("created_at DESC").Limit(limit).Offset(offset).
		Find(&returns).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return returns, total, nil
}

func (s *returnService) PerformTransition(
	ctx context.Context,
	returnID uint,
	event, note, actorRole string,
	actorID *uint,
) (*workflow.TransitionResult, error) {
	if s.engine == nil {
		return nil, utils.ErrInternal(errors.New("workflow engine not configured"))
	}
	result, err := s.engine.Transition(ctx, workflow.TransitionRequest{
		WorkflowKey: constants.WorkflowEntityReturn,
		EntityID:    returnID,
		Event:       event,
		ActorID:     actorID,
		ActorRole:   actorRole,
		Note:        note,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
