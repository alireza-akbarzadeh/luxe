package controllers

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/services/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type WorkflowController struct {
	svc      services.WorkflowServiceInterface
	validate *validator.Validate
}

func NewWorkflowController(svc services.WorkflowServiceInterface) *WorkflowController {
	return &WorkflowController{svc: svc, validate: validator.New()}
}

// ─── Generic (authenticated) endpoints ────────────────────────────────────────

// GetDefinition returns a workflow definition with states and colors for the frontend.
// @Summary      Get workflow definition
// @Description  Returns a workflow with its states (including color codes) and transitions.
// @Tags         Workflows
// @Produce      json
// @Security     BearerAuth
// @Param        key path string true "Workflow key (order, product, shipment, return, user)"
// @Success      200 {object} utils.Response{data=dto.WorkflowDefinitionView}
// @Failure      404 {object} utils.Response
// @Router       /workflows/{key} [get]
func (ctrl *WorkflowController) GetDefinition(c *gin.Context) {
	key := c.Param("key")
	wf, err := ctrl.svc.GetDefinition(c.Request.Context(), key)
	if err != nil {
		RespondServiceError(c, err, "failed to load workflow")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, toWorkflowDefinitionView(wf))
}

// PerformTransition applies a transition event to an entity instance.
// @Summary      Perform a workflow transition
// @Description  Moves an entity to a new state by firing an event. Records who changed what, when.
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        key      path string true "Workflow key"
// @Param        entityId path int    true "Entity ID"
// @Param        request  body dto.PerformTransitionRequest true "Transition event"
// @Success      200 {object} utils.Response{data=dto.TransitionResultView}
// @Failure      400 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /workflows/{key}/{entityId}/transition [post]
func (ctrl *WorkflowController) PerformTransition(c *gin.Context) {
	key := c.Param("key")
	entityID, ok := parseUintParam(c, "entityId")
	if !ok {
		return
	}

	var req dto.PerformTransitionRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	actorID, _ := middleware.GetUserID(c)
	actorRole, _ := middleware.GetUserRole(c)
	var actorIDPtr *uint
	if actorID != 0 {
		actorIDPtr = &actorID
	}

	result, err := ctrl.svc.Engine().Transition(c.Request.Context(), workflow.TransitionRequest{
		WorkflowKey: key,
		EntityID:    entityID,
		Event:       req.Event,
		ActorID:     actorIDPtr,
		ActorRole:   actorRole,
		Note:        req.Note,
		Metadata:    req.Metadata,
	})
	if err != nil {
		RespondServiceError(c, err, "failed to perform transition")
		return
	}
	utils.SuccessResponse(c, "transition applied", toTransitionResultView(result))
}

// AvailableTransitions lists the actions allowed from the entity's current state.
// @Summary      List available transitions
// @Tags         Workflows
// @Produce      json
// @Security     BearerAuth
// @Param        key      path string true "Workflow key"
// @Param        entityId path int    true "Entity ID"
// @Success      200 {object} utils.Response
// @Router       /workflows/{key}/{entityId}/available-transitions [get]
func (ctrl *WorkflowController) AvailableTransitions(c *gin.Context) {
	key := c.Param("key")
	entityID, ok := parseUintParam(c, "entityId")
	if !ok {
		return
	}
	current, transitions, err := ctrl.svc.Engine().AvailableTransitions(c.Request.Context(), key, entityID)
	if err != nil {
		RespondServiceError(c, err, "failed to load available transitions")
		return
	}
	views := make([]dto.TransitionView, 0, len(transitions))
	for i := range transitions {
		views = append(views, toTransitionView(&transitions[i]))
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"current_state": toStateView(current),
		"transitions":   views,
	})
}

// History returns the audit trail for an entity instance.
// @Summary      Get workflow history
// @Description  Returns who changed the entity's state, from/to (with colors), when, and any note.
// @Tags         Workflows
// @Produce      json
// @Security     BearerAuth
// @Param        key      path string true "Workflow key / entity type (order, product, shipment, return, user)"
// @Param        entityId path int    true "Entity ID"
// @Param        limit    query int    false "Items per page"
// @Param        offset   query int    false "Offset"
// @Success      200 {object} utils.Response
// @Router       /workflows/{key}/{entityId}/history [get]
func (ctrl *WorkflowController) History(c *gin.Context) {
	entityType := c.Param("key")
	entityID, ok := parseUintParam(c, "entityId")
	if !ok {
		return
	}
	limit, offset := paginationParams(c, constants.DefaultLimit)
	logs, total, err := ctrl.svc.Engine().History(c.Request.Context(), entityType, entityID, limit, offset)
	if err != nil {
		RespondServiceError(c, err, "failed to load workflow history")
		return
	}
	entries := make([]dto.WorkflowHistoryEntry, 0, len(logs))
	for i := range logs {
		entries = append(entries, toHistoryEntry(&logs[i]))
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"history": entries,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// ─── Admin CRUD endpoints ─────────────────────────────────────────────────────

// ListWorkflows returns all workflow definitions.
// @Summary      List workflows (admin)
// @Tags         Workflows
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response
// @Router       /admin/workflows [get]
func (ctrl *WorkflowController) ListWorkflows(c *gin.Context) {
	workflows, err := ctrl.svc.ListWorkflows(c.Request.Context())
	if err != nil {
		RespondServiceError(c, err, "failed to list workflows")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, workflows)
}

// CreateWorkflow creates a new workflow definition.
// @Summary      Create workflow (admin)
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateWorkflowRequest true "Workflow"
// @Success      201 {object} utils.Response
// @Router       /admin/workflows [post]
func (ctrl *WorkflowController) CreateWorkflow(c *gin.Context) {
	var req dto.CreateWorkflowRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	wf, err := ctrl.svc.CreateWorkflow(c.Request.Context(), req)
	if err != nil {
		RespondServiceError(c, err, "failed to create workflow")
		return
	}
	utils.CreatedResponse(c, "workflow created", wf)
}

// UpdateWorkflow updates a workflow definition.
// @Summary      Update workflow (admin)
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Workflow ID"
// @Param        request body dto.UpdateWorkflowRequest true "Fields"
// @Success      200 {object} utils.Response
// @Router       /admin/workflows/{id} [patch]
func (ctrl *WorkflowController) UpdateWorkflow(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateWorkflowRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	wf, err := ctrl.svc.UpdateWorkflow(c.Request.Context(), id, req)
	if err != nil {
		RespondServiceError(c, err, "failed to update workflow")
		return
	}
	utils.SuccessResponse(c, "workflow updated", wf)
}

// DeleteWorkflow soft-deletes a workflow definition.
// @Summary      Delete workflow (admin)
// @Tags         Workflows
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Workflow ID"
// @Success      200 {object} utils.Response
// @Router       /admin/workflows/{id} [delete]
func (ctrl *WorkflowController) DeleteWorkflow(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := ctrl.svc.DeleteWorkflow(c.Request.Context(), id); err != nil {
		RespondServiceError(c, err, "failed to delete workflow")
		return
	}
	utils.SuccessResponse(c, "workflow deleted", nil)
}

// CreateState adds a state (with color) to a workflow.
// @Summary      Create workflow state (admin)
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Workflow ID"
// @Param        request body dto.CreateWorkflowStateRequest true "State"
// @Success      201 {object} utils.Response
// @Router       /admin/workflows/{id}/states [post]
func (ctrl *WorkflowController) CreateState(c *gin.Context) {
	workflowID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.CreateWorkflowStateRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	state, err := ctrl.svc.CreateState(c.Request.Context(), workflowID, req)
	if err != nil {
		RespondServiceError(c, err, "failed to create state")
		return
	}
	utils.CreatedResponse(c, "state created", state)
}

// UpdateState updates a state's properties (name, color, flags).
// @Summary      Update workflow state (admin)
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Workflow ID"
// @Param        stateId path int true "State ID"
// @Param        request body dto.UpdateWorkflowStateRequest true "Fields"
// @Success      200 {object} utils.Response
// @Router       /admin/workflows/{id}/states/{stateId} [patch]
func (ctrl *WorkflowController) UpdateState(c *gin.Context) {
	stateID, ok := parseUintParam(c, "stateId")
	if !ok {
		return
	}
	var req dto.UpdateWorkflowStateRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	state, err := ctrl.svc.UpdateState(c.Request.Context(), stateID, req)
	if err != nil {
		RespondServiceError(c, err, "failed to update state")
		return
	}
	utils.SuccessResponse(c, "state updated", state)
}

// DeleteState removes a state.
// @Summary      Delete workflow state (admin)
// @Tags         Workflows
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Workflow ID"
// @Param        stateId path int true "State ID"
// @Success      200 {object} utils.Response
// @Router       /admin/workflows/{id}/states/{stateId} [delete]
func (ctrl *WorkflowController) DeleteState(c *gin.Context) {
	stateID, ok := parseUintParam(c, "stateId")
	if !ok {
		return
	}
	if err := ctrl.svc.DeleteState(c.Request.Context(), stateID); err != nil {
		RespondServiceError(c, err, "failed to delete state")
		return
	}
	utils.SuccessResponse(c, "state deleted", nil)
}

// CreateTransition adds a transition rule to a workflow.
// @Summary      Create workflow transition (admin)
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Workflow ID"
// @Param        request body dto.CreateWorkflowTransitionRequest true "Transition"
// @Success      201 {object} utils.Response
// @Router       /admin/workflows/{id}/transitions [post]
func (ctrl *WorkflowController) CreateTransition(c *gin.Context) {
	workflowID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.CreateWorkflowTransitionRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	trans, err := ctrl.svc.CreateTransition(c.Request.Context(), workflowID, req)
	if err != nil {
		RespondServiceError(c, err, "failed to create transition")
		return
	}
	utils.CreatedResponse(c, "transition created", trans)
}

// UpdateTransition updates a transition rule.
// @Summary      Update workflow transition (admin)
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path int true "Workflow ID"
// @Param        transitionId path int true "Transition ID"
// @Param        request body dto.UpdateWorkflowTransitionRequest true "Fields"
// @Success      200 {object} utils.Response
// @Router       /admin/workflows/{id}/transitions/{transitionId} [patch]
func (ctrl *WorkflowController) UpdateTransition(c *gin.Context) {
	transitionID, ok := parseUintParam(c, "transitionId")
	if !ok {
		return
	}
	var req dto.UpdateWorkflowTransitionRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	trans, err := ctrl.svc.UpdateTransition(c.Request.Context(), transitionID, req)
	if err != nil {
		RespondServiceError(c, err, "failed to update transition")
		return
	}
	utils.SuccessResponse(c, "transition updated", trans)
}

// DeleteTransition removes a transition rule.
// @Summary      Delete workflow transition (admin)
// @Tags         Workflows
// @Produce      json
// @Security     BearerAuth
// @Param        id           path int true "Workflow ID"
// @Param        transitionId path int true "Transition ID"
// @Success      200 {object} utils.Response
// @Router       /admin/workflows/{id}/transitions/{transitionId} [delete]
func (ctrl *WorkflowController) DeleteTransition(c *gin.Context) {
	transitionID, ok := parseUintParam(c, "transitionId")
	if !ok {
		return
	}
	if err := ctrl.svc.DeleteTransition(c.Request.Context(), transitionID); err != nil {
		RespondServiceError(c, err, "failed to delete transition")
		return
	}
	utils.SuccessResponse(c, "transition deleted", nil)
}

// ─── view mappers ─────────────────────────────────────────────────────────────

func toStateView(s *models.WorkflowState) *dto.StateView {
	if s == nil {
		return nil
	}
	return &dto.StateView{
		ID:        s.ID,
		Code:      s.Code,
		Name:      s.Name,
		Color:     s.Color,
		TextColor: s.TextColor,
		IsInitial: s.IsInitial,
		IsFinal:   s.IsFinal,
		SortOrder: s.SortOrder,
	}
}

func toTransitionView(t *models.WorkflowTransition) dto.TransitionView {
	return dto.TransitionView{
		Event:        t.Event,
		Name:         t.Name,
		RequiredRole: t.RequiredRole,
		ToState:      toStateView(t.ToState),
	}
}

func toWorkflowDefinitionView(wf *models.Workflow) dto.WorkflowDefinitionView {
	states := make([]dto.StateView, 0, len(wf.States))
	for i := range wf.States {
		if sv := toStateView(&wf.States[i]); sv != nil {
			states = append(states, *sv)
		}
	}
	transitions := make([]dto.TransitionView, 0, len(wf.Transitions))
	for i := range wf.Transitions {
		transitions = append(transitions, toTransitionView(&wf.Transitions[i]))
	}
	return dto.WorkflowDefinitionView{
		ID:          wf.ID,
		Key:         wf.Key,
		Name:        wf.Name,
		Description: wf.Description,
		EntityType:  wf.EntityType,
		IsActive:    wf.IsActive,
		States:      states,
		Transitions: transitions,
	}
}

func toTransitionResultView(r *workflow.TransitionResult) dto.TransitionResultView {
	return dto.TransitionResultView{
		EntityType: r.EntityType,
		EntityID:   r.EntityID,
		Event:      r.Event,
		FromState:  toStateView(r.From),
		ToState:    toStateView(r.To),
	}
}

func toHistoryEntry(l *models.WorkflowTransitionLog) dto.WorkflowHistoryEntry {
	entry := dto.WorkflowHistoryEntry{
		ID:        l.ID,
		Event:     l.Event,
		FromState: toStateView(l.FromState),
		ToState:   toStateView(l.ToState),
		UserID:    l.UserID,
		Note:      l.Note,
		Success:   l.Success,
		ErrorMsg:  l.ErrorMsg,
		CreatedAt: l.CreatedAt,
	}
	if l.User != nil {
		entry.UserName = l.User.FirstName + " " + l.User.LastName
	}
	return entry
}
