package workflow

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	infraworkflow "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Queries orchestrates workflow definition read use cases.
type Queries struct {
	repo *postgres.WorkflowRepository
}

// NewQueries creates workflow query use cases.
func NewQueries(repo *postgres.WorkflowRepository) *Queries {
	return &Queries{repo: repo}
}

// GetDefinition loads a workflow definition by key.
func (q *Queries) GetDefinition(ctx context.Context, key string) (*models.Workflow, error) {
	wf, err := q.repo.FindByKey(ctx, key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("workflow not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return wf, nil
}

// ListWorkflows returns all workflow definitions.
func (q *Queries) ListWorkflows(ctx context.Context) ([]models.Workflow, error) {
	workflows, err := q.repo.ListAll(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return workflows, nil
}

// Commands orchestrates workflow definition write use cases.
type Commands struct {
	repo *postgres.WorkflowRepository
}

// NewCommands creates workflow command use cases.
func NewCommands(repo *postgres.WorkflowRepository) *Commands {
	return &Commands{repo: repo}
}

// CreateWorkflow inserts a workflow definition.
func (c *Commands) CreateWorkflow(ctx context.Context, req dto.CreateWorkflowRequest) (*models.Workflow, error) {
	wf := models.Workflow{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		EntityType:  req.EntityType,
		IsActive:    true,
	}
	if err := c.repo.CreateWorkflow(ctx, &wf); err != nil {
		if postgres.IsUniqueViolation(err) {
			return nil, utils.ErrConflict("a workflow with key '" + req.Key + "' already exists")
		}
		return nil, utils.ErrInternal(err)
	}
	return &wf, nil
}

// UpdateWorkflow updates a workflow definition.
func (c *Commands) UpdateWorkflow(ctx context.Context, id uint, req dto.UpdateWorkflowRequest) (*models.Workflow, error) {
	wf, err := c.repo.FindWorkflowByID(ctx, id)
	if err != nil {
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
	if err := c.repo.SaveWorkflow(ctx, wf); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return wf, nil
}

// DeleteWorkflow removes a workflow definition.
func (c *Commands) DeleteWorkflow(ctx context.Context, id uint) error {
	if err := c.repo.DeleteWorkflow(ctx, id); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// CreateState inserts a workflow state.
func (c *Commands) CreateState(ctx context.Context, workflowID uint, req dto.CreateWorkflowStateRequest) (*models.WorkflowState, error) {
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
	if err := c.repo.CreateState(ctx, &state); err != nil {
		if postgres.IsUniqueViolation(err) {
			return nil, utils.ErrConflict("state code '" + req.Code + "' already exists in this workflow")
		}
		return nil, utils.ErrInternal(err)
	}
	return &state, nil
}

// UpdateState updates a workflow state.
func (c *Commands) UpdateState(ctx context.Context, stateID uint, req dto.UpdateWorkflowStateRequest) (*models.WorkflowState, error) {
	state, err := c.repo.FindStateByID(ctx, stateID)
	if err != nil {
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
	if err := c.repo.SaveState(ctx, state); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return state, nil
}

// DeleteState removes a workflow state.
func (c *Commands) DeleteState(ctx context.Context, stateID uint) error {
	if err := c.repo.DeleteState(ctx, stateID); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// CreateTransition inserts a workflow transition.
func (c *Commands) CreateTransition(ctx context.Context, workflowID uint, req dto.CreateWorkflowTransitionRequest) (*models.WorkflowTransition, error) {
	toState, err := c.stateByCode(ctx, workflowID, req.ToStateCode)
	if err != nil {
		return nil, err
	}

	var fromStateID *uint
	if req.FromStateCode != "" {
		fromState, err := c.stateByCode(ctx, workflowID, req.FromStateCode)
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
	if err := c.repo.CreateTransition(ctx, &trans); err != nil {
		if postgres.IsUniqueViolation(err) {
			return nil, utils.ErrConflict("a transition for event '" + req.Event + "' from this state already exists")
		}
		return nil, utils.ErrInternal(err)
	}
	return &trans, nil
}

// UpdateTransition updates a workflow transition.
func (c *Commands) UpdateTransition(ctx context.Context, transitionID uint, req dto.UpdateWorkflowTransitionRequest) (*models.WorkflowTransition, error) {
	trans, err := c.repo.FindTransitionByID(ctx, transitionID)
	if err != nil {
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
	if err := c.repo.SaveTransition(ctx, trans); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return trans, nil
}

// DeleteTransition removes a workflow transition.
func (c *Commands) DeleteTransition(ctx context.Context, transitionID uint) error {
	if err := c.repo.DeleteTransition(ctx, transitionID); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (c *Commands) stateByCode(ctx context.Context, workflowID uint, code string) (*models.WorkflowState, error) {
	state, err := c.repo.FindStateByCode(ctx, workflowID, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrBadRequest("state code '" + code + "' not found in workflow")
		}
		return nil, utils.ErrInternal(err)
	}
	return state, nil
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

// EngineHolder exposes the infrastructure workflow engine.
type EngineHolder struct {
	engine *infraworkflow.Engine
}

// NewEngineHolder wraps the workflow engine.
func NewEngineHolder(engine *infraworkflow.Engine) *EngineHolder {
	return &EngineHolder{engine: engine}
}

// Engine returns the underlying engine.
func (h *EngineHolder) Engine() *infraworkflow.Engine {
	return h.engine
}
