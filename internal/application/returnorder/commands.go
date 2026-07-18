package returnorder

import (
	"context"
	"errors"

	appmembership "github.com/alireza-akbarzadeh/luxe/internal/application/membership"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	domainreturn "github.com/alireza-akbarzadeh/luxe/internal/domain/returnorder"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

var openReturnExcludedStatuses = []string{"rejected", "closed", "refunded"}

// Commands orchestrates return write use cases.
type Commands struct {
	repo       *postgres.ReturnRepository
	engine     *workflow.Engine
	membership *appmembership.Service
}

// NewCommands creates return command use cases.
func NewCommands(repo *postgres.ReturnRepository, engine *workflow.Engine, membership *appmembership.Service) *Commands {
	return &Commands{repo: repo, engine: engine, membership: membership}
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

	existing, err := c.repo.CountOpenForOrder(ctx, req.OrderID, userID, openReturnExcludedStatuses)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	if err := domainreturn.ValidateCreateRequest(order.Status, existing); err != nil {
		switch {
		case errors.Is(err, domainreturn.ErrOrderNotEligible):
			return nil, utils.ErrBadRequest(err.Error())
		case errors.Is(err, domainreturn.ErrOpenReturnExists):
			return nil, utils.ErrConflict(err.Error())
		default:
			return nil, utils.ErrInternal(err)
		}
	}

	windowDays := constants.FreeReturnWindowDays
	if c.membership != nil {
		if days, err := c.membership.UserReturnWindowDays(ctx, userID); err == nil {
			windowDays = days
		}
	}
	if err := domainreturn.ValidateReturnWindow(order.UpdatedAt, windowDays); err != nil {
		if errors.Is(err, domainreturn.ErrReturnWindowExpired) {
			return nil, utils.ErrBadRequest(err.Error())
		}
		return nil, utils.ErrInternal(err)
	}

	returnType := req.ReturnType
	if returnType == "" {
		returnType = dto.ReturnTypeRefund
	}
	if err := domainreturn.ValidateReturnType(returnType); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}

	refundAmount := order.TotalAmount
	if returnType == dto.ReturnTypeExchange {
		refundAmount = 0
	}

	ret := &models.Return{
		OrderID:       req.OrderID,
		UserID:        userID,
		Reason:        req.Reason,
		Status:        "requested",
		ReturnType:    returnType,
		RefundAmount:  refundAmount,
		ExchangeNotes: req.ExchangeNotes,
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

// UpdateNotes saves admin-only notes on a return request.
func (c *Commands) UpdateNotes(ctx context.Context, returnID uint, notes string) (*models.Return, error) {
	ret, err := c.repo.FindByID(ctx, returnID, 0, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("return not found")
		}
		return nil, utils.ErrInternal(err)
	}

	ret.AdminNotes = notes
	if err := c.repo.Update(ctx, ret); err != nil {
		return nil, utils.ErrInternal(err)
	}

	return c.repo.FindByID(ctx, returnID, 0, true)
}
