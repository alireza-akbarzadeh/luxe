package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Workflow is a runtime-configurable state machine definition for an entity type.
type Workflow struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Key         string         `gorm:"uniqueIndex;not null" json:"key"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description,omitempty"`
	EntityType  string         `gorm:"not null" json:"entity_type"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	States      []WorkflowState      `gorm:"foreignKey:WorkflowID" json:"states,omitempty"`
	Transitions []WorkflowTransition `gorm:"foreignKey:WorkflowID" json:"transitions,omitempty"`
}

// WorkflowState is a single state within a workflow, including frontend color codes.
type WorkflowState struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	WorkflowID  uint      `gorm:"not null;index;uniqueIndex:uq_wf_state_code" json:"workflow_id"`
	Code        string    `gorm:"not null;uniqueIndex:uq_wf_state_code" json:"code"`
	Name        string    `gorm:"not null" json:"name"`
	Color       string    `gorm:"not null;default:'#6B7280'" json:"color"`
	TextColor   string    `gorm:"column:text_color;not null;default:'#FFFFFF'" json:"text_color"`
	Description string    `json:"description,omitempty"`
	IsInitial   bool      `gorm:"not null;default:false" json:"is_initial"`
	IsFinal     bool      `gorm:"not null;default:false" json:"is_final"`
	SortOrder   int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WorkflowTransition defines a legal move between two states, triggered by an event.
// A nil FromStateID is a wildcard meaning the transition applies from any state.
type WorkflowTransition struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	WorkflowID   uint      `gorm:"not null;index" json:"workflow_id"`
	FromStateID  *uint     `gorm:"index" json:"from_state_id,omitempty"`
	ToStateID    uint      `gorm:"not null" json:"to_state_id"`
	Event        string    `gorm:"not null" json:"event"`
	Name         string    `gorm:"not null" json:"name"`
	RequiredRole string    `json:"required_role,omitempty"`
	GuardKey     string    `json:"guard_key,omitempty"`
	HookKey      string    `json:"hook_key,omitempty"`
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
	SortOrder    int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	FromState *WorkflowState `gorm:"foreignKey:FromStateID" json:"from_state,omitempty"`
	ToState   *WorkflowState `gorm:"foreignKey:ToStateID" json:"to_state,omitempty"`
}

// WorkflowTransitionLog is the immutable audit/history record of a state change.
type WorkflowTransitionLog struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	WorkflowID   uint           `gorm:"not null;index" json:"workflow_id"`
	EntityType   string         `gorm:"not null;index:idx_wf_logs_entity,priority:1" json:"entity_type"`
	EntityID     uint           `gorm:"not null;index:idx_wf_logs_entity,priority:2" json:"entity_id"`
	FromStateID  *uint          `json:"from_state_id,omitempty"`
	ToStateID    *uint          `json:"to_state_id,omitempty"`
	TransitionID *uint          `json:"transition_id,omitempty"`
	Event        string         `gorm:"not null" json:"event"`
	UserID       *uint          `gorm:"index" json:"user_id,omitempty"`
	Note         string         `json:"note,omitempty"`
	Success      bool           `gorm:"not null;default:true" json:"success"`
	ErrorMsg     string         `gorm:"column:error_msg" json:"error_msg,omitempty"`
	Metadata     datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`

	FromState *WorkflowState `gorm:"foreignKey:FromStateID" json:"from_state,omitempty"`
	ToState   *WorkflowState `gorm:"foreignKey:ToStateID" json:"to_state,omitempty"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
