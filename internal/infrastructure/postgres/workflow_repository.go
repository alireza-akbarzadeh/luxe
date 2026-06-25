package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// WorkflowRepository persists workflow definitions with GORM.
type WorkflowRepository struct {
	db *gorm.DB
}

// NewWorkflowRepository creates a GORM-backed workflow repository.
func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

// FindByKey loads a workflow definition with states and transitions.
func (r *WorkflowRepository) FindByKey(ctx context.Context, key string) (*models.Workflow, error) {
	var wf models.Workflow
	err := r.db.WithContext(ctx).
		Preload("States", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order") }).
		Preload("Transitions", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order") }).
		Preload("Transitions.FromState").
		Preload("Transitions.ToState").
		Where("key = ?", key).
		First(&wf).Error
	if err != nil {
		return nil, err
	}
	return &wf, nil
}

// ListAll returns all workflows ordered by key.
func (r *WorkflowRepository) ListAll(ctx context.Context) ([]models.Workflow, error) {
	var workflows []models.Workflow
	err := r.db.WithContext(ctx).Order("key").Find(&workflows).Error
	return workflows, err
}

// CreateWorkflow inserts a workflow definition.
func (r *WorkflowRepository) CreateWorkflow(ctx context.Context, wf *models.Workflow) error {
	return r.db.WithContext(ctx).Create(wf).Error
}

// FindWorkflowByID loads a workflow by id.
func (r *WorkflowRepository) FindWorkflowByID(ctx context.Context, id uint) (*models.Workflow, error) {
	var wf models.Workflow
	if err := r.db.WithContext(ctx).First(&wf, id).Error; err != nil {
		return nil, err
	}
	return &wf, nil
}

// SaveWorkflow persists workflow changes.
func (r *WorkflowRepository) SaveWorkflow(ctx context.Context, wf *models.Workflow) error {
	return r.db.WithContext(ctx).Save(wf).Error
}

// DeleteWorkflow removes a workflow by id.
func (r *WorkflowRepository) DeleteWorkflow(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Workflow{}, id).Error
}

// CreateState inserts a workflow state.
func (r *WorkflowRepository) CreateState(ctx context.Context, state *models.WorkflowState) error {
	return r.db.WithContext(ctx).Create(state).Error
}

// FindStateByID loads a workflow state.
func (r *WorkflowRepository) FindStateByID(ctx context.Context, id uint) (*models.WorkflowState, error) {
	var state models.WorkflowState
	if err := r.db.WithContext(ctx).First(&state, id).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

// SaveState persists workflow state changes.
func (r *WorkflowRepository) SaveState(ctx context.Context, state *models.WorkflowState) error {
	return r.db.WithContext(ctx).Save(state).Error
}

// DeleteState removes a workflow state.
func (r *WorkflowRepository) DeleteState(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.WorkflowState{}, id).Error
}

// FindStateByCode loads a state by workflow id and code.
func (r *WorkflowRepository) FindStateByCode(ctx context.Context, workflowID uint, code string) (*models.WorkflowState, error) {
	var state models.WorkflowState
	err := r.db.WithContext(ctx).
		Where("workflow_id = ? AND code = ?", workflowID, code).
		First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// CreateTransition inserts a workflow transition.
func (r *WorkflowRepository) CreateTransition(ctx context.Context, trans *models.WorkflowTransition) error {
	return r.db.WithContext(ctx).Create(trans).Error
}

// FindTransitionByID loads a workflow transition.
func (r *WorkflowRepository) FindTransitionByID(ctx context.Context, id uint) (*models.WorkflowTransition, error) {
	var trans models.WorkflowTransition
	if err := r.db.WithContext(ctx).First(&trans, id).Error; err != nil {
		return nil, err
	}
	return &trans, nil
}

// SaveTransition persists transition changes.
func (r *WorkflowRepository) SaveTransition(ctx context.Context, trans *models.WorkflowTransition) error {
	return r.db.WithContext(ctx).Save(trans).Error
}

// DeleteTransition removes a workflow transition.
func (r *WorkflowRepository) DeleteTransition(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.WorkflowTransition{}, id).Error
}
