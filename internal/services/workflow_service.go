package services

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/services/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

// WorkflowServiceInterface manages workflow definitions and exposes the runtime engine.
type WorkflowServiceInterface interface {
	// Definition queries
	GetDefinition(ctx context.Context, key string) (*models.Workflow, error)
	ListWorkflows(ctx context.Context) ([]models.Workflow, error)

	// Workflow CRUD
	CreateWorkflow(ctx context.Context, req dto.CreateWorkflowRequest) (*models.Workflow, error)
	UpdateWorkflow(ctx context.Context, id uint, req dto.UpdateWorkflowRequest) (*models.Workflow, error)
	DeleteWorkflow(ctx context.Context, id uint) error

	// State CRUD
	CreateState(ctx context.Context, workflowID uint, req dto.CreateWorkflowStateRequest) (*models.WorkflowState, error)
	UpdateState(ctx context.Context, stateID uint, req dto.UpdateWorkflowStateRequest) (*models.WorkflowState, error)
	DeleteState(ctx context.Context, stateID uint) error

	// Transition CRUD
	CreateTransition(ctx context.Context, workflowID uint, req dto.CreateWorkflowTransitionRequest) (*models.WorkflowTransition, error)
	UpdateTransition(ctx context.Context, transitionID uint, req dto.UpdateWorkflowTransitionRequest) (*models.WorkflowTransition, error)
	DeleteTransition(ctx context.Context, transitionID uint) error

	// Runtime
	Engine() *workflow.Engine
}

type workflowService struct {
	db     *gorm.DB
	engine *workflow.Engine
}

func NewWorkflowService(db *gorm.DB, engine *workflow.Engine) WorkflowServiceInterface {
	return &workflowService{db: db, engine: engine}
}

func (s *workflowService) Engine() *workflow.Engine { return s.engine }

func (s *workflowService) GetDefinition(ctx context.Context, key string) (*models.Workflow, error) {
	var wf models.Workflow
	err := s.db.WithContext(ctx).
		Preload("States", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order") }).
		Preload("Transitions", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order") }).
		Preload("Transitions.FromState").
		Preload("Transitions.ToState").
		Where("key = ?", key).
		First(&wf).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("workflow not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &wf, nil
}

func (s *workflowService) ListWorkflows(ctx context.Context) ([]models.Workflow, error) {
	var workflows []models.Workflow
	if err := s.db.WithContext(ctx).Order("key").Find(&workflows).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return workflows, nil
}

func (s *workflowService) CreateWorkflow(ctx context.Context, req dto.CreateWorkflowRequest) (*models.Workflow, error) {
	wf := models.Workflow{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		EntityType:  req.EntityType,
		IsActive:    true,
	}
	if err := s.db.WithContext(ctx).Create(&wf).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, utils.ErrConflict("a workflow with key '" + req.Key + "' already exists")
		}
		return nil, utils.ErrInternal(err)
	}
	return &wf, nil
}

func (s *workflowService) UpdateWorkflow(ctx context.Context, id uint, req dto.UpdateWorkflowRequest) (*models.Workflow, error) {
	var wf models.Workflow
	if err := s.db.WithContext(ctx).First(&wf, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("workflow not found")
		}
		return nil, utils.ErrInternal(err)
	}
	if req.Name != nil {
		wf.Name = *req.Name
	}
	if req.Description != nil {
		wf.Description = *req.Description
	}
	if req.IsActive != nil {
		wf.IsActive = *req.IsActive
	}
	if err := s.db.WithContext(ctx).Save(&wf).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &wf, nil
}

func (s *workflowService) DeleteWorkflow(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Delete(&models.Workflow{}, id).Error; err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (s *workflowService) CreateState(ctx context.Context, workflowID uint, req dto.CreateWorkflowStateRequest) (*models.WorkflowState, error) {
	state := models.WorkflowState{
		WorkflowID:  workflowID,
		Code:        req.Code,
		Name:        req.Name,
		Color:       defaultStr(req.Color, "#6B7280"),
		TextColor:   defaultStr(req.TextColor, "#FFFFFF"),
		Description: req.Description,
		IsInitial:   req.IsInitial,
		IsFinal:     req.IsFinal,
		SortOrder:   req.SortOrder,
	}
	if err := s.db.WithContext(ctx).Create(&state).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, utils.ErrConflict("state code '" + req.Code + "' already exists in this workflow")
		}
		return nil, utils.ErrInternal(err)
	}
	return &state, nil
}

func (s *workflowService) UpdateState(ctx context.Context, stateID uint, req dto.UpdateWorkflowStateRequest) (*models.WorkflowState, error) {
	var state models.WorkflowState
	if err := s.db.WithContext(ctx).First(&state, stateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("state not found")
		}
		return nil, utils.ErrInternal(err)
	}
	if req.Name != nil {
		state.Name = *req.Name
	}
	if req.Color != nil {
		state.Color = *req.Color
	}
	if req.TextColor != nil {
		state.TextColor = *req.TextColor
	}
	if req.Description != nil {
		state.Description = *req.Description
	}
	if req.IsInitial != nil {
		state.IsInitial = *req.IsInitial
	}
	if req.IsFinal != nil {
		state.IsFinal = *req.IsFinal
	}
	if req.SortOrder != nil {
		state.SortOrder = *req.SortOrder
	}
	if err := s.db.WithContext(ctx).Save(&state).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &state, nil
}

func (s *workflowService) DeleteState(ctx context.Context, stateID uint) error {
	if err := s.db.WithContext(ctx).Delete(&models.WorkflowState{}, stateID).Error; err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (s *workflowService) CreateTransition(ctx context.Context, workflowID uint, req dto.CreateWorkflowTransitionRequest) (*models.WorkflowTransition, error) {
	toState, err := s.stateByCode(ctx, workflowID, req.ToStateCode)
	if err != nil {
		return nil, err
	}

	var fromStateID *uint
	if req.FromStateCode != "" {
		fromState, err := s.stateByCode(ctx, workflowID, req.FromStateCode)
		if err != nil {
			return nil, err
		}
		fromStateID = &fromState.ID
	}

	trans := models.WorkflowTransition{
		WorkflowID:   workflowID,
		FromStateID:  fromStateID,
		ToStateID:    toState.ID,
		Event:        req.Event,
		Name:         req.Name,
		RequiredRole: req.RequiredRole,
		GuardKey:     req.GuardKey,
		HookKey:      req.HookKey,
		IsActive:     true,
		SortOrder:    req.SortOrder,
	}
	if err := s.db.WithContext(ctx).Create(&trans).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, utils.ErrConflict("a transition for event '" + req.Event + "' from this state already exists")
		}
		return nil, utils.ErrInternal(err)
	}
	return &trans, nil
}

func (s *workflowService) UpdateTransition(ctx context.Context, transitionID uint, req dto.UpdateWorkflowTransitionRequest) (*models.WorkflowTransition, error) {
	var trans models.WorkflowTransition
	if err := s.db.WithContext(ctx).First(&trans, transitionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("transition not found")
		}
		return nil, utils.ErrInternal(err)
	}
	if req.Name != nil {
		trans.Name = *req.Name
	}
	if req.RequiredRole != nil {
		trans.RequiredRole = *req.RequiredRole
	}
	if req.GuardKey != nil {
		trans.GuardKey = *req.GuardKey
	}
	if req.HookKey != nil {
		trans.HookKey = *req.HookKey
	}
	if req.IsActive != nil {
		trans.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		trans.SortOrder = *req.SortOrder
	}
	if err := s.db.WithContext(ctx).Save(&trans).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &trans, nil
}

func (s *workflowService) DeleteTransition(ctx context.Context, transitionID uint) error {
	if err := s.db.WithContext(ctx).Delete(&models.WorkflowTransition{}, transitionID).Error; err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (s *workflowService) stateByCode(ctx context.Context, workflowID uint, code string) (*models.WorkflowState, error) {
	var state models.WorkflowState
	err := s.db.WithContext(ctx).
		Where("workflow_id = ? AND code = ?", workflowID, code).
		First(&state).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrBadRequest("state code '" + code + "' not found in workflow")
		}
		return nil, utils.ErrInternal(err)
	}
	return &state, nil
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
