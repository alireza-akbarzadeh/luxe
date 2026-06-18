// Package workflow implements a generic, DB-driven state-machine engine.
// Workflow definitions (states, transitions, colors, guards, hooks) live in the
// database and are editable at runtime; the engine validates and applies
// transitions, records immutable history, and runs side-effect hooks.
package workflow

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// GuardFunc returns a non-nil error to veto a transition (e.g. payment not settled).
type GuardFunc func(ctx context.Context, entityID uint) error

// HookFunc runs after a transition commits (notifications, job enqueue, emails).
type HookFunc func(ctx context.Context, entityID uint, meta map[string]interface{}) error

// entityConfig maps an entity type to its table and whether it mirrors state into
// a legacy "status" string column for backward compatibility.
type entityConfig struct {
	table        string
	statusMirror bool
	activeMirror bool // mirrors is_active from workflow state code ("active" => true)
}

var entityConfigs = map[string]entityConfig{
	constants.WorkflowEntityOrder:    {table: "orders", statusMirror: true},
	constants.WorkflowEntityProduct:  {table: "products", statusMirror: true},
	constants.WorkflowEntityShipment: {table: "shipments", statusMirror: true},
	constants.WorkflowEntityReturn:   {table: "returns", statusMirror: true},
	constants.WorkflowEntityUser:     {table: "users", statusMirror: false},
	constants.WorkflowEntityCategory:  {table: "categories", activeMirror: true},
}

// statusMirrorMaps translate a workflow state code into the legacy status string.
// Missing entries fall back to the state code itself (identity).
var statusMirrorMaps = map[string]map[string]string{
	constants.WorkflowEntityOrder: {
		"created": "pending", "pending_payment": "pending", "packed": "processing",
	},
	constants.WorkflowEntityProduct: {
		"under_review": "inactive", "approved": "inactive", "published": "active",
		"out_of_stock": "active", "discontinued": "inactive",
	},
	constants.WorkflowEntityShipment: {
		"ready_for_pickup": "processing", "picked_up": "processing", "in_transit": "shipped",
		"out_for_delivery": "shipped", "failed_delivery": "shipped", "returned": "cancelled",
	},
}

func mirrorStatus(entityType, stateCode string) string {
	if m, ok := statusMirrorMaps[entityType]; ok {
		if mapped, ok := m[stateCode]; ok {
			return mapped
		}
	}
	return stateCode
}

func applyEntityMirrors(cfg entityConfig, entityType, stateCode string, updates map[string]interface{}) {
	if cfg.statusMirror {
		updates["status"] = mirrorStatus(entityType, stateCode)
	}
	if cfg.activeMirror {
		updates["is_active"] = stateCode == "active"
	}
}

// Engine is the runtime state-machine processor.
type Engine struct {
	db     *gorm.DB
	guards map[string]GuardFunc
	hooks  map[string]HookFunc
}

func NewEngine(db *gorm.DB) *Engine {
	return &Engine{
		db:     db,
		guards: make(map[string]GuardFunc),
		hooks:  make(map[string]HookFunc),
	}
}

// RegisterGuard wires a guard function under the key referenced by transition rows.
func (e *Engine) RegisterGuard(key string, fn GuardFunc) { e.guards[key] = fn }

// RegisterHook wires a post-transition side-effect under the key referenced by transition rows.
func (e *Engine) RegisterHook(key string, fn HookFunc) { e.hooks[key] = fn }

// TransitionRequest describes an attempt to move an entity to a new state.
type TransitionRequest struct {
	WorkflowKey string
	EntityID    uint
	Event       string
	ActorID     *uint
	ActorRole   string
	Note        string
	Metadata    map[string]interface{}
}

// TransitionResult is returned on a successful transition.
type TransitionResult struct {
	WorkflowKey string
	EntityType  string
	EntityID    uint
	Event       string
	From        *models.WorkflowState
	To          *models.WorkflowState
}

// Transition validates and applies a state change, recording history and running hooks.
func (e *Engine) Transition(ctx context.Context, req TransitionRequest) (*TransitionResult, error) {
	wf, err := e.loadWorkflow(ctx, req.WorkflowKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := entityConfigs[wf.EntityType]
	if !ok {
		return nil, utils.ErrInternal(errors.New("no entity config for " + wf.EntityType))
	}

	currentStateID, err := e.currentStateID(ctx, cfg.table, req.EntityID)
	if err != nil {
		return nil, err
	}

	fromState, err := e.resolveFromState(ctx, wf.ID, currentStateID)
	if err != nil {
		return nil, err
	}

	trans, err := e.findTransition(ctx, wf.ID, fromState.ID, req.Event)
	if err != nil {
		return nil, err
	}

	if err := e.checkRole(trans, req.ActorRole); err != nil {
		e.logFailure(ctx, wf, req, &fromState.ID, &trans.ToStateID, &trans.ID, err)
		return nil, err
	}

	if trans.GuardKey != "" {
		if guard, ok := e.guards[trans.GuardKey]; ok {
			if gErr := guard(ctx, req.EntityID); gErr != nil {
				e.logFailure(ctx, wf, req, &fromState.ID, &trans.ToStateID, &trans.ID, gErr)
				return nil, gErr
			}
		}
	}

	var toState models.WorkflowState
	if err := e.db.WithContext(ctx).First(&toState, trans.ToStateID).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	meta := e.marshalMeta(req.Metadata)
	fromID := fromState.ID

	err = e.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{"workflow_state_id": trans.ToStateID}
		applyEntityMirrors(cfg, wf.EntityType, toState.Code, updates)
		if uErr := tx.Table(cfg.table).Where("id = ?", req.EntityID).Updates(updates).Error; uErr != nil {
			return uErr
		}

		logEntry := models.WorkflowTransitionLog{
			WorkflowID:   wf.ID,
			EntityType:   wf.EntityType,
			EntityID:     req.EntityID,
			FromStateID:  &fromID,
			ToStateID:    &trans.ToStateID,
			TransitionID: &trans.ID,
			Event:        req.Event,
			UserID:       req.ActorID,
			Note:         req.Note,
			Success:      true,
			Metadata:     meta,
		}
		return tx.Create(&logEntry).Error
	})
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	// Post-transition side effects (best-effort; failures are logged, not rolled back).
	if trans.HookKey != "" {
		if hook, ok := e.hooks[trans.HookKey]; ok {
			if hErr := hook(ctx, req.EntityID, req.Metadata); hErr != nil {
				utils.Log.WithError(hErr).
					WithField("hook", trans.HookKey).
					WithField("entity_id", req.EntityID).
					Warn("workflow hook failed")
				e.logFailure(ctx, wf, req, &fromID, &trans.ToStateID, &trans.ID, hErr)
			}
		}
	}

	return &TransitionResult{
		WorkflowKey: wf.Key,
		EntityType:  wf.EntityType,
		EntityID:    req.EntityID,
		Event:       req.Event,
		From:        fromState,
		To:          &toState,
	}, nil
}

// SetStateRequest is an admin/system-driven state override that bypasses transition
// validation and guards but still records full audit history.
type SetStateRequest struct {
	WorkflowKey     string
	EntityID        uint
	TargetStateCode string
	Event           string // audit label, e.g. "admin_set" or "system_sync"
	ActorID         *uint
	ActorRole       string
	Note            string
	Metadata        map[string]interface{}
}

// SetState forces an entity to a target state by code, recording who changed it.
// Use for admin overrides and internal synchronization; use Transition for
// validated, event-driven moves.
func (e *Engine) SetState(ctx context.Context, req SetStateRequest) (*TransitionResult, error) {
	wf, err := e.loadWorkflow(ctx, req.WorkflowKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := entityConfigs[wf.EntityType]
	if !ok {
		return nil, utils.ErrInternal(errors.New("no entity config for " + wf.EntityType))
	}

	var toState models.WorkflowState
	err = e.db.WithContext(ctx).
		Where("workflow_id = ? AND code = ?", wf.ID, req.TargetStateCode).
		First(&toState).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrBadRequest("unknown target state '" + req.TargetStateCode + "'")
		}
		return nil, utils.ErrInternal(err)
	}

	currentStateID, err := e.currentStateID(ctx, cfg.table, req.EntityID)
	if err != nil {
		return nil, err
	}

	event := req.Event
	if event == "" {
		event = "set_state"
	}

	err = e.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{"workflow_state_id": toState.ID}
		applyEntityMirrors(cfg, wf.EntityType, toState.Code, updates)
		if uErr := tx.Table(cfg.table).Where("id = ?", req.EntityID).Updates(updates).Error; uErr != nil {
			return uErr
		}
		logEntry := models.WorkflowTransitionLog{
			WorkflowID:  wf.ID,
			EntityType:  wf.EntityType,
			EntityID:    req.EntityID,
			FromStateID: currentStateID,
			ToStateID:   &toState.ID,
			Event:       event,
			UserID:      req.ActorID,
			Note:        req.Note,
			Success:     true,
			Metadata:    e.marshalMeta(req.Metadata),
		}
		return tx.Create(&logEntry).Error
	})
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	var fromState *models.WorkflowState
	if currentStateID != nil {
		var fs models.WorkflowState
		if e.db.WithContext(ctx).First(&fs, *currentStateID).Error == nil {
			fromState = &fs
		}
	}
	return &TransitionResult{
		WorkflowKey: wf.Key,
		EntityType:  wf.EntityType,
		EntityID:    req.EntityID,
		Event:       event,
		From:        fromState,
		To:          &toState,
	}, nil
}

// AvailableTransitions returns the transitions a caller can attempt from the entity's current state.
func (e *Engine) AvailableTransitions(ctx context.Context, workflowKey string, entityID uint) (*models.WorkflowState, []models.WorkflowTransition, error) {
	wf, err := e.loadWorkflow(ctx, workflowKey)
	if err != nil {
		return nil, nil, err
	}
	cfg, ok := entityConfigs[wf.EntityType]
	if !ok {
		return nil, nil, utils.ErrInternal(errors.New("no entity config for " + wf.EntityType))
	}
	currentStateID, err := e.currentStateID(ctx, cfg.table, entityID)
	if err != nil {
		return nil, nil, err
	}
	fromState, err := e.resolveFromState(ctx, wf.ID, currentStateID)
	if err != nil {
		return nil, nil, err
	}

	var transitions []models.WorkflowTransition
	if err := e.db.WithContext(ctx).
		Preload("ToState").
		Where("workflow_id = ? AND is_active = ?", wf.ID, true).
		Where("from_state_id = ? OR from_state_id IS NULL", fromState.ID).
		Order("sort_order").
		Find(&transitions).Error; err != nil {
		return nil, nil, utils.ErrInternal(err)
	}
	return fromState, transitions, nil
}

// History returns the audit trail for an entity, newest first.
func (e *Engine) History(ctx context.Context, entityType string, entityID uint, limit, offset int) ([]models.WorkflowTransitionLog, int64, error) {
	q := e.db.WithContext(ctx).Model(&models.WorkflowTransitionLog{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	var logs []models.WorkflowTransitionLog
	if err := q.Preload("FromState").Preload("ToState").Preload("User").
		Order("created_at DESC").Limit(limit).Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return logs, total, nil
}

// ─── internals ────────────────────────────────────────────────────────────────

func (e *Engine) loadWorkflow(ctx context.Context, key string) (*models.Workflow, error) {
	var wf models.Workflow
	err := e.db.WithContext(ctx).Where("key = ? AND is_active = ?", key, true).First(&wf).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("workflow '" + key + "' not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &wf, nil
}

func (e *Engine) currentStateID(ctx context.Context, table string, entityID uint) (*uint, error) {
	var result struct{ WorkflowStateID *uint }
	tx := e.db.WithContext(ctx).Table(table).
		Select("workflow_state_id").
		Where("id = ?", entityID).
		Scan(&result)
	if tx.Error != nil {
		return nil, utils.ErrInternal(tx.Error)
	}
	if tx.RowsAffected == 0 {
		return nil, utils.ErrNotFound("entity not found")
	}
	return result.WorkflowStateID, nil
}

// resolveFromState returns the entity's current state, or the workflow's initial state when unset.
func (e *Engine) resolveFromState(ctx context.Context, workflowID uint, currentStateID *uint) (*models.WorkflowState, error) {
	var state models.WorkflowState
	if currentStateID != nil {
		if err := e.db.WithContext(ctx).First(&state, *currentStateID).Error; err != nil {
			return nil, utils.ErrInternal(err)
		}
		return &state, nil
	}
	err := e.db.WithContext(ctx).
		Where("workflow_id = ? AND is_initial = ?", workflowID, true).
		First(&state).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrInternal(errors.New("workflow has no initial state"))
		}
		return nil, utils.ErrInternal(err)
	}
	return &state, nil
}

// findTransition resolves the transition for (event, current state), preferring a
// state-specific rule over a wildcard ("any state") rule.
func (e *Engine) findTransition(ctx context.Context, workflowID, fromStateID uint, event string) (*models.WorkflowTransition, error) {
	var trans models.WorkflowTransition
	err := e.db.WithContext(ctx).
		Where("workflow_id = ? AND event = ? AND is_active = ?", workflowID, event, true).
		Where("from_state_id = ? OR from_state_id IS NULL", fromStateID).
		Order("from_state_id IS NULL"). // false (specific) sorts before true (wildcard)
		First(&trans).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrBadRequest("transition '" + event + "' is not allowed from the current state")
		}
		return nil, utils.ErrInternal(err)
	}
	return &trans, nil
}

// checkRole enforces the transition's required role. Admins may perform any transition.
func (e *Engine) checkRole(trans *models.WorkflowTransition, actorRole string) error {
	if trans.RequiredRole == "" {
		return nil
	}
	if actorRole == constants.RoleAdmin || actorRole == trans.RequiredRole {
		return nil
	}
	return utils.ErrForbidden("you do not have permission to perform '" + trans.Event + "'")
}

func (e *Engine) marshalMeta(meta map[string]interface{}) datatypes.JSON {
	if len(meta) == 0 {
		return nil
	}
	b, err := json.Marshal(meta)
	if err != nil {
		return nil
	}
	return datatypes.JSON(b)
}

func (e *Engine) logFailure(ctx context.Context, wf *models.Workflow, req TransitionRequest, fromID, toID, transID *uint, cause error) {
	entry := models.WorkflowTransitionLog{
		WorkflowID:   wf.ID,
		EntityType:   wf.EntityType,
		EntityID:     req.EntityID,
		FromStateID:  fromID,
		ToStateID:    toID,
		TransitionID: transID,
		Event:        req.Event,
		UserID:       req.ActorID,
		Note:         req.Note,
		Success:      false,
		ErrorMsg:     cause.Error(),
		Metadata:     e.marshalMeta(req.Metadata),
	}
	if err := e.db.WithContext(ctx).Create(&entry).Error; err != nil {
		utils.Log.WithError(err).Warn("failed to write workflow failure log")
	}
}
