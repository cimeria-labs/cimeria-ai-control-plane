package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

type LeadScoreRuleResponse struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	EventType   string `json:"event_type"`
	Weight      int32  `json:"weight"`
	MaxPerEmail *int32 `json:"max_per_email"`
	Enabled     bool   `json:"enabled"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type UpsertLeadScoreRuleRequest struct {
	EventType   string `json:"event_type"`
	Weight      int32  `json:"weight"`
	MaxPerEmail *int32 `json:"max_per_email"`
	Enabled     *bool  `json:"enabled"`
}

var validLeadScoreEvents = map[string]bool{
	"opened": true, "clicked": true, "replied": true, "forwarded": true,
	"bounced": true, "complained": true, "unsubscribed": true,
}

func leadScoreRuleToResponse(rule db.LeadScoreRule) LeadScoreRuleResponse {
	var max *int32
	if rule.MaxPerEmail.Valid {
		value := rule.MaxPerEmail.Int32
		max = &value
	}
	return LeadScoreRuleResponse{
		ID:          uuidToString(rule.ID),
		WorkspaceID: uuidToString(rule.WorkspaceID),
		EventType:   rule.EventType,
		Weight:      rule.Weight,
		MaxPerEmail: max,
		Enabled:     rule.Enabled,
		CreatedAt:   timestampToString(rule.CreatedAt),
		UpdatedAt:   timestampToString(rule.UpdatedAt),
	}
}

func (h *Handler) ListLeadScoreRules(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	rules, err := h.Queries.ListLeadScoreRules(r.Context(), parseUUID(workspaceID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list lead score rules")
		return
	}
	resp := make([]LeadScoreRuleResponse, 0, len(rules))
	for _, rule := range rules {
		resp = append(resp, leadScoreRuleToResponse(rule))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) UpsertLeadScoreRule(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	var req UpsertLeadScoreRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	eventType := strings.TrimSpace(req.EventType)
	if eventType == "" {
		eventType = strings.TrimSpace(chi.URLParam(r, "eventType"))
	}
	if !validLeadScoreEvents[eventType] {
		writeError(w, http.StatusBadRequest, "invalid event type")
		return
	}
	maxPerEmail := pgtype.Int4{}
	if req.MaxPerEmail != nil {
		maxPerEmail = pgtype.Int4{Int32: *req.MaxPerEmail, Valid: true}
	}
	enabled := pgtype.Bool{}
	if req.Enabled != nil {
		enabled = pgtype.Bool{Bool: *req.Enabled, Valid: true}
	}
	rule, err := h.Queries.UpsertLeadScoreRule(r.Context(), db.UpsertLeadScoreRuleParams{
		WorkspaceID: parseUUID(workspaceID),
		EventType:   eventType,
		Weight:      req.Weight,
		MaxPerEmail: maxPerEmail,
		Enabled:     enabled,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save lead score rule")
		return
	}
	writeJSON(w, http.StatusOK, leadScoreRuleToResponse(rule))
}

func (h *Handler) DeleteLeadScoreRule(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	eventType := strings.TrimSpace(chi.URLParam(r, "eventType"))
	rule, err := h.Queries.GetLeadScoreRule(r.Context(), db.GetLeadScoreRuleParams{
		WorkspaceID: parseUUID(workspaceID),
		EventType:   eventType,
	})
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "lead score rule not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get lead score rule")
		return
	}
	if err := h.Queries.DeleteLeadScoreRule(r.Context(), rule.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete lead score rule")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
