package privacyrule

import (
	"context"
	"errors"

	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	domainprivacyrule "github.com/alireza-akbarzadeh/luxe/internal/domain/privacyrule"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service handles privacy rule HTTP-oriented use cases.
type Service struct {
	engine   *workflow.Engine
	commands *Commands
	queries  *Queries
}

// NewService wires privacy rule application use cases.
func NewService(db *gorm.DB, engine *workflow.Engine) *Service {
	repo := postgres.NewPrivacyRuleRepository(db)
	return &Service{
		engine:   engine,
		commands: NewCommands(domainprivacyrule.NewService(), repo),
		queries:  NewQueries(repo),
	}
}

func (s *Service) syncWorkflow(ctx context.Context, ruleID uint, status string) {
	if !appworkflow.ApplyPrivacyRuleWorkflow(ctx, s.engine, ruleID, status, nil) {
		utils.Log.WithField("privacy_rule_id", ruleID).WithField("status", status).
			Debug("privacy rule workflow sync skipped or failed")
	}
}

// Create creates a privacy rule and syncs workflow.
func (s *Service) Create(ctx context.Context, req *dto.CreatePrivacyRuleRequest) (*dto.PrivacyRuleResponse, error) {
	rule, err := s.commands.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	s.syncWorkflow(ctx, rule.ID, rule.Status)

	loaded, err := s.queries.GetByID(ctx, rule.ID)
	if err != nil {
		return nil, err
	}
	return ToResponse(loaded), nil
}

// GetByID returns a privacy rule by ID.
func (s *Service) GetByID(ctx context.Context, id uint) (*dto.PrivacyRuleResponse, error) {
	rule, err := s.queries.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("privacy rule not found")
		}
		return nil, err
	}
	return ToResponse(rule), nil
}

// GetActiveByKey returns an active privacy rule for apps to parse.
func (s *Service) GetActiveByKey(ctx context.Context, key, locale string) (*dto.PrivacyRuleResponse, error) {
	rule, err := s.queries.GetActiveByKey(ctx, key, locale)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("privacy rule not found")
		}
		return nil, err
	}
	return ToResponse(rule), nil
}

// List returns paginated privacy rules (admin).
func (s *Service) List(ctx context.Context, req *dto.ListPrivacyRulesRequest) ([]dto.PrivacyRuleResponse, int64, error) {
	items, total, err := s.queries.List(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.PrivacyRuleResponse, 0, len(items))
	for i := range items {
		resp = append(resp, *ToResponse(&items[i]))
	}
	return resp, total, nil
}

// ListActivePublic returns active rules for storefront/app consumption.
func (s *Service) ListActivePublic(ctx context.Context, req *dto.ListPrivacyRulesRequest) ([]dto.PrivacyRuleResponse, int64, error) {
	status := "active"
	req.Status = status
	return s.List(ctx, req)
}

// ListActiveByProvider returns active rules for a provider.
func (s *Service) ListActiveByProvider(ctx context.Context, provider, locale string) ([]dto.PrivacyRuleResponse, error) {
	items, err := s.queries.ListActiveByProvider(ctx, provider, locale)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.PrivacyRuleResponse, 0, len(items))
	for i := range items {
		resp = append(resp, *ToResponse(&items[i]))
	}
	return resp, nil
}

// Update updates a privacy rule and optionally syncs workflow.
func (s *Service) Update(ctx context.Context, id uint, req *dto.UpdatePrivacyRuleRequest) (*dto.PrivacyRuleResponse, error) {
	rule, err := s.queries.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("privacy rule not found")
		}
		return nil, err
	}

	if err := s.commands.Update(ctx, rule, req); err != nil {
		return nil, err
	}

	if req.Status != nil {
		s.syncWorkflow(ctx, id, *req.Status)
	}

	loaded, err := s.queries.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ToResponse(loaded), nil
}

// Delete removes a privacy rule.
func (s *Service) Delete(ctx context.Context, id uint) error {
	rows, err := s.commands.Delete(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return utils.ErrNotFound("privacy rule not found")
	}
	return nil
}
