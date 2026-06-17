package dto

import "time"

// ─── Workflow definition CRUD ─────────────────────────────────────────────────

type CreateWorkflowRequest struct {
	Key         string `json:"key"         validate:"required,min=2,max=64"`
	Name        string `json:"name"        validate:"required,min=2,max=128"`
	Description string `json:"description" validate:"omitempty,max=512"`
	EntityType  string `json:"entity_type" validate:"required,min=2,max=64"`
}

type UpdateWorkflowRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=2,max=128"`
	Description *string `json:"description" validate:"omitempty,max=512"`
	IsActive    *bool   `json:"is_active"`
}

// ─── State CRUD ───────────────────────────────────────────────────────────────

type CreateWorkflowStateRequest struct {
	Code        string `json:"code"        validate:"required,min=1,max=64"`
	Name        string `json:"name"        validate:"required,min=1,max=128"`
	Color       string `json:"color"       validate:"omitempty,hexcolor"`
	TextColor   string `json:"text_color"  validate:"omitempty,hexcolor"`
	Description string `json:"description" validate:"omitempty,max=512"`
	IsInitial   bool   `json:"is_initial"`
	IsFinal     bool   `json:"is_final"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateWorkflowStateRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=1,max=128"`
	Color       *string `json:"color"       validate:"omitempty,hexcolor"`
	TextColor   *string `json:"text_color"  validate:"omitempty,hexcolor"`
	Description *string `json:"description" validate:"omitempty,max=512"`
	IsInitial   *bool   `json:"is_initial"`
	IsFinal     *bool   `json:"is_final"`
	SortOrder   *int    `json:"sort_order"`
}

// ─── Transition CRUD ──────────────────────────────────────────────────────────

type CreateWorkflowTransitionRequest struct {
	FromStateCode string `json:"from_state_code" validate:"omitempty,max=64"` // empty = wildcard "any state"
	ToStateCode   string `json:"to_state_code"   validate:"required,max=64"`
	Event         string `json:"event"           validate:"required,min=1,max=64"`
	Name          string `json:"name"            validate:"required,min=1,max=128"`
	RequiredRole  string `json:"required_role"   validate:"omitempty,max=32"`
	GuardKey      string `json:"guard_key"       validate:"omitempty,max=64"`
	HookKey       string `json:"hook_key"        validate:"omitempty,max=64"`
	SortOrder     int    `json:"sort_order"`
}

type UpdateWorkflowTransitionRequest struct {
	Name         *string `json:"name"          validate:"omitempty,min=1,max=128"`
	RequiredRole *string `json:"required_role" validate:"omitempty,max=32"`
	GuardKey     *string `json:"guard_key"     validate:"omitempty,max=64"`
	HookKey      *string `json:"hook_key"      validate:"omitempty,max=64"`
	IsActive     *bool   `json:"is_active"`
	SortOrder    *int    `json:"sort_order"`
}

// ─── Perform transition ───────────────────────────────────────────────────────

type PerformTransitionRequest struct {
	Event    string                 `json:"event"    validate:"required,min=1,max=64"`
	Note     string                 `json:"note"     validate:"omitempty,max=512"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ─── Responses ────────────────────────────────────────────────────────────────

// StateView is the frontend-facing representation of a state with its colors.
type StateView struct {
	ID        uint   `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	TextColor string `json:"text_color"`
	IsInitial bool   `json:"is_initial"`
	IsFinal   bool   `json:"is_final"`
	SortOrder int    `json:"sort_order"`
}

// TransitionView describes one available action a caller may perform.
type TransitionView struct {
	Event        string     `json:"event"`
	Name         string     `json:"name"`
	RequiredRole string     `json:"required_role,omitempty"`
	ToState      *StateView `json:"to_state,omitempty"`
}

// WorkflowDefinitionView is the full definition the frontend needs to render badges.
type WorkflowDefinitionView struct {
	ID          uint             `json:"id"`
	Key         string           `json:"key"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	EntityType  string           `json:"entity_type"`
	IsActive    bool             `json:"is_active"`
	States      []StateView      `json:"states"`
	Transitions []TransitionView `json:"transitions"`
}

// TransitionResultView is returned after a successful transition.
type TransitionResultView struct {
	EntityType string     `json:"entity_type"`
	EntityID   uint       `json:"entity_id"`
	Event      string     `json:"event"`
	FromState  *StateView `json:"from_state,omitempty"`
	ToState    *StateView `json:"to_state"`
}

// WorkflowHistoryEntry is one audit row: who moved the entity from which state to which.
type WorkflowHistoryEntry struct {
	ID         uint       `json:"id"`
	Event      string     `json:"event"`
	FromState  *StateView `json:"from_state,omitempty"`
	ToState    *StateView `json:"to_state,omitempty"`
	UserID     *uint      `json:"user_id,omitempty"`
	UserName   string     `json:"user_name,omitempty"`
	Note       string     `json:"note,omitempty"`
	Success    bool       `json:"success"`
	ErrorMsg   string     `json:"error_msg,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
