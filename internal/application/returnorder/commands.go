package returnorder

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

var returnEligibleOrderStatuses = map[string]bool{
	constants.OrderStatusDelivered: true,
	"completed":                    true,
}

var openReturnExcludedStatuses = []string{"rejected", "closed", "refunded"}

// Commands orchestrates return write use cases.
type Commands struct {
	repo   *postgres.ReturnRepository
	engine *workflow.Engine
}

// NewCommands creates return command use cases.
func NewCommands(repo *postgres.ReturnRepository, engine *workflow.Engine) *Commands {
	return &Commands{repo: repo, engine: engine}
}

// Create inserts a return request and syncs workflow state.
func (c *Commands) Create(ctx context.Context, userID uint, req dto.CreateReturnRequest) (*models.Return, error) {
	order, err := c.repo.FindUserOrder(ctx, req.OrderID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("order not found")
		}
		return nil, utils.ErrInternal(err)
	}

	if !returnEligibleOrderStatuses[order.Status] {
		return nil, utils.ErrBadRequest("returns are only allowed for delivered or completed orders")
	}

	existing, err := c.repo.CountOpenForOrder(ctx, req.OrderID, userID, openReturnExcludedStatuses)
	if err != nil {
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

	if err := c.repo.Create(ctx, ret); err != nil {
		return nil, utils.ErrInternal(err)
	}

	if c.engine != nil {
		_, _ = c.engine.SetState(ctx, workflow.SetStateRequest{
			WorkflowKey:     constants.WorkflowEntityReturn,
			EntityID:        ret.ID,
			TargetStateCode: "requested",
			Event:           "return_requested",
			ActorID:         &userID,
		})
	}
	return ret, nil
}

// PerformTransition applies a workflow event to a return.
func (c *Commands) PerformTransition(
	ctx context.Context,
	returnID uint,
	event, note, actorRole string,
	actorID *uint,
) (*workflow.TransitionResult, error) {
	if c.engine == nil {
		return nil, utils.ErrInternal(errors.New("workflow engine not configured"))
	}
	return c.engine.Transition(ctx, workflow.TransitionRequest{
		WorkflowKey: constants.WorkflowEntityReturn,
		EntityID:    returnID,
		Event:       event,
		ActorID:     actorID,
		ActorRole:   actorRole,
		Note:        note,
	})
}
