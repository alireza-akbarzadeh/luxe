package review

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

// Commands orchestrates review write use cases.
type Commands struct {
	repo   *postgres.ReviewRepository
	engine *workflow.Engine
}

// NewCommands creates review command use cases.
func NewCommands(repo *postgres.ReviewRepository, engine *workflow.Engine) *Commands {
	return &Commands{repo: repo, engine: engine}
}

func (c *Commands) syncWorkflow(ctx context.Context, reviewID, userID uint, stateCode, event string) {
	if c.engine == nil {
		return
	}
	_, _ = c.engine.SetState(ctx, workflow.SetStateRequest{
		WorkflowKey:     constants.WorkflowEntityReview,
		EntityID:        reviewID,
		TargetStateCode: stateCode,
		Event:           event,
		ActorID:         &userID,
	})
}

func (c *Commands) isApprovedStatus(status string) bool {
	return status == "approved"
}

// UpdateProductStats recalculates product rating from approved reviews.
func (c *Commands) UpdateProductStats(ctx context.Context, productID uint) {
	avg, count, err := c.repo.ProductStats(ctx, productID)
	if err != nil {
		return
	}
	_ = c.repo.UpdateProductRating(ctx, productID, avg, count)
}

// Create inserts a review and enqueues moderation workflow.
func (c *Commands) Create(ctx context.Context, userID uint, req dto.CreateReviewRequest) (*models.Review, error) {
	_, err := c.repo.FindByUserAndProduct(ctx, userID, req.ProductID)
	if err == nil {
		return nil, utils.ErrBadRequest("you have already reviewed this product")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrInternal(err)
	}

	review := &models.Review{
		ProductID: req.ProductID,
		UserID:    userID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		Title:     req.Title,
	}
	if err := c.repo.Create(ctx, review); err != nil {
		return nil, utils.ErrInternal(err)
	}

	c.syncWorkflow(ctx, review.ID, userID, "pending", "submitted")
	return c.repo.FindByID(ctx, review.ID)
}

// Update updates a review and resets workflow when content changes.
func (c *Commands) Update(ctx context.Context, userID, reviewID uint, req dto.UpdateReviewRequest) (*models.Review, error) {
	review, err := c.repo.FindByID(ctx, reviewID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("review not found")
		}
		return nil, utils.ErrInternal(err)
	}
	if review.UserID != userID {
		return nil, utils.ErrNotFound("review not found")
	}

	previousStatus := review.Status
	contentChanged := false

	if req.Rating != nil {
		review.Rating = *req.Rating
		contentChanged = true
	}
	if req.Comment != nil {
		review.Comment = *req.Comment
		contentChanged = true
	}
	if req.Title != nil {
		review.Title = *req.Title
		contentChanged = true
	}

	if err := c.repo.Save(ctx, review); err != nil {
		return nil, utils.ErrInternal(err)
	}

	if contentChanged && previousStatus != "pending" {
		c.syncWorkflow(ctx, review.ID, userID, "pending", "resubmitted")
	}

	if c.isApprovedStatus(previousStatus) || contentChanged {
		c.UpdateProductStats(ctx, review.ProductID)
	}

	return c.repo.FindByID(ctx, review.ID)
}

// Delete removes a review and recalculates stats when needed.
func (c *Commands) Delete(ctx context.Context, userID, reviewID uint) error {
	review, err := c.repo.FindByIDForUser(ctx, reviewID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("review not found")
		}
		return utils.ErrInternal(err)
	}

	wasApproved := c.isApprovedStatus(review.Status)
	productID := review.ProductID

	if err := c.repo.Delete(ctx, review); err != nil {
		return utils.ErrInternal(err)
	}

	if wasApproved {
		c.UpdateProductStats(ctx, productID)
	}
	return nil
}

// PerformTransition applies a workflow event and updates product stats.
func (c *Commands) PerformTransition(
	ctx context.Context,
	reviewID uint,
	event, note, actorRole string,
	actorID *uint,
) (*workflow.TransitionResult, error) {
	if c.engine == nil {
		return nil, utils.ErrInternal(errors.New("workflow engine not configured"))
	}

	review, err := c.repo.FindByIDFields(ctx, reviewID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("review not found")
		}
		return nil, utils.ErrInternal(err)
	}
	previousApproved := c.isApprovedStatus(review.Status)

	result, err := c.engine.Transition(ctx, workflow.TransitionRequest{
		WorkflowKey: constants.WorkflowEntityReview,
		EntityID:    reviewID,
		Event:       event,
		ActorID:     actorID,
		ActorRole:   actorRole,
		Note:        note,
	})
	if err != nil {
		return nil, err
	}

	if previousApproved || result.To.Code == "approved" {
		c.UpdateProductStats(ctx, review.ProductID)
	}

	return result, nil
}
