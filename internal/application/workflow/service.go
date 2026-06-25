package workflow

import (
	"context"

	domainworkflow "github.com/alireza-akbarzadeh/luxe/internal/domain/workflow"
	infraworkflow "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
)

// Service exposes workflow use cases to application handlers and orchestrators.
type Service struct {
	engine *infraworkflow.Engine
}

// NewService wraps the infrastructure workflow engine.
func NewService(engine *infraworkflow.Engine) *Service {
	return &Service{engine: engine}
}

// Engine returns the underlying engine for guards, hooks, and full API access.
func (s *Service) Engine() *infraworkflow.Engine {
	return s.engine
}

// Transition runs a validated workflow transition.
func (s *Service) Transition(ctx context.Context, req domainworkflow.TransitionRequest) (*domainworkflow.TransitionResult, error) {
	result, err := s.engine.Transition(ctx, infraworkflow.TransitionRequest{
		WorkflowKey: req.WorkflowKey,
		EntityID:    req.EntityID,
		Event:       req.Event,
		ActorID:     req.ActorID,
		ActorRole:   req.ActorRole,
		Note:        req.Note,
		Metadata:    req.Metadata,
	})
	if err != nil {
		return nil, err
	}
	return mapTransitionResult(result), nil
}

func mapTransitionResult(r *infraworkflow.TransitionResult) *domainworkflow.TransitionResult {
	if r == nil {
		return nil
	}
	out := &domainworkflow.TransitionResult{
		WorkflowKey: r.WorkflowKey,
		EntityType:  r.EntityType,
		EntityID:    r.EntityID,
		Event:       r.Event,
	}
	if r.From != nil {
		out.From = &domainworkflow.State{ID: r.From.ID, Code: r.From.Code, Name: r.From.Name, Color: r.From.Color}
	}
	if r.To != nil {
		out.To = &domainworkflow.State{ID: r.To.ID, Code: r.To.Code, Name: r.To.Name, Color: r.To.Color}
	}
	return out
}
