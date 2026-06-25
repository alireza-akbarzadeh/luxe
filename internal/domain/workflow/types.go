// Package workflow defines the workflow domain port and value types.
package workflow

import "context"

// State is a workflow state snapshot without persistence details.
type State struct {
	ID    uint
	Code  string
	Name  string
	Color string
}

// TransitionRequest describes an event-driven state change.
type TransitionRequest struct {
	WorkflowKey string
	EntityID    uint
	Event       string
	ActorID     *uint
	ActorRole   string
	Note        string
	Metadata    map[string]interface{}
}

// TransitionResult is returned after a successful transition.
type TransitionResult struct {
	WorkflowKey string
	EntityType  string
	EntityID    uint
	Event       string
	From        *State
	To          *State
}

// EnginePort is the domain-facing workflow engine contract.
// Implemented by infrastructure/workflow.Engine.
type EnginePort interface {
	Transition(ctx context.Context, req TransitionRequest) (*TransitionResult, error)
}
