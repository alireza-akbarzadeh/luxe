package services

import (
	"context"

	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// WorkflowServiceInterface manages workflow definitions and exposes the runtime engine.
type WorkflowServiceInterface interface {
	GetDefinition(ctx context.Context, key string) (*models.Workflow, error)
	ListWorkflows(ctx context.Context) ([]models.Workflow, error)

	CreateWorkflow(ctx context.Context, req dto.CreateWorkflowRequest) (*models.Workflow, error)
	UpdateWorkflow(ctx context.Context, id uint, req dto.UpdateWorkflowRequest) (*models.Workflow, error)
	DeleteWorkflow(ctx context.Context, id uint) error

	CreateState(ctx context.Context, workflowID uint, req dto.CreateWorkflowStateRequest) (*models.WorkflowState, error)
	UpdateState(ctx context.Context, stateID uint, req dto.UpdateWorkflowStateRequest) (*models.WorkflowState, error)
	DeleteState(ctx context.Context, stateID uint) error

	CreateTransition(ctx context.Context, workflowID uint, req dto.CreateWorkflowTransitionRequest) (*models.WorkflowTransition, error)
	UpdateTransition(ctx context.Context, transitionID uint, req dto.UpdateWorkflowTransitionRequest) (*models.WorkflowTransition, error)
	DeleteTransition(ctx context.Context, transitionID uint) error

	Engine() *workflow.Engine
}

type workflowService struct {
	queries  *appworkflow.Queries
	commands *appworkflow.Commands
	engine   *workflow.Engine
}

func NewWorkflowService(db *gorm.DB, engine *workflow.Engine) WorkflowServiceInterface {
	repo := postgres.NewWorkflowRepository(db)
	return &workflowService{
		queries:  appworkflow.NewQueries(repo),
		commands: appworkflow.NewCommands(repo),
		engine:   engine,
	}
}

func (s *workflowService) Engine() *workflow.Engine { return s.engine }

func (s *workflowService) GetDefinition(ctx context.Context, key string) (*models.Workflow, error) {
	return s.queries.GetDefinition(ctx, key)
}

func (s *workflowService) ListWorkflows(ctx context.Context) ([]models.Workflow, error) {
	return s.queries.ListWorkflows(ctx)
}

func (s *workflowService) CreateWorkflow(ctx context.Context, req dto.CreateWorkflowRequest) (*models.Workflow, error) {
	return s.commands.CreateWorkflow(ctx, req)
}

func (s *workflowService) UpdateWorkflow(ctx context.Context, id uint, req dto.UpdateWorkflowRequest) (*models.Workflow, error) {
	return s.commands.UpdateWorkflow(ctx, id, req)
}

func (s *workflowService) DeleteWorkflow(ctx context.Context, id uint) error {
	return s.commands.DeleteWorkflow(ctx, id)
}

func (s *workflowService) CreateState(ctx context.Context, workflowID uint, req dto.CreateWorkflowStateRequest) (*models.WorkflowState, error) {
	return s.commands.CreateState(ctx, workflowID, req)
}

func (s *workflowService) UpdateState(ctx context.Context, stateID uint, req dto.UpdateWorkflowStateRequest) (*models.WorkflowState, error) {
	return s.commands.UpdateState(ctx, stateID, req)
}

func (s *workflowService) DeleteState(ctx context.Context, stateID uint) error {
	return s.commands.DeleteState(ctx, stateID)
}

func (s *workflowService) CreateTransition(ctx context.Context, workflowID uint, req dto.CreateWorkflowTransitionRequest) (*models.WorkflowTransition, error) {
	return s.commands.CreateTransition(ctx, workflowID, req)
}

func (s *workflowService) UpdateTransition(ctx context.Context, transitionID uint, req dto.UpdateWorkflowTransitionRequest) (*models.WorkflowTransition, error) {
	return s.commands.UpdateTransition(ctx, transitionID, req)
}

func (s *workflowService) DeleteTransition(ctx context.Context, transitionID uint) error {
	return s.commands.DeleteTransition(ctx, transitionID)
}
