package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

// ---------------------------------------------------------------------------
// LeadSource
// ---------------------------------------------------------------------------

type LeadSourceResponse struct {
	ID                  string         `json:"id"`
	WorkspaceID         string         `json:"workspace_id"`
	Name                string         `json:"name"`
	Slug                string         `json:"slug"`
	Provider            string         `json:"provider"`
	Config              map[string]any `json:"config"`
	IsActive            bool           `json:"is_active"`
	AutoApprove         bool           `json:"auto_approve"`
	EnrichmentEnabled   bool           `json:"enrichment_enabled"`
	CreatedAt           string         `json:"created_at"`
	UpdatedAt           string         `json:"updated_at"`
}

func leadSourceToResponse(s db.LeadSource) LeadSourceResponse {
	config := make(map[string]any)
	if len(s.Config) > 0 {
		_ = json.Unmarshal(s.Config, &config)
	}
	return LeadSourceResponse{
		ID:                uuidToString(s.ID),
		WorkspaceID:       uuidToString(s.WorkspaceID),
		Name:              s.Name,
		Slug:              s.Slug,
		Provider:          s.Provider,
		Config:            config,
		IsActive:          s.IsActive,
		AutoApprove:       s.AutoApprove,
		EnrichmentEnabled: s.EnrichmentEnabled,
		CreatedAt:         timestampToString(s.CreatedAt),
		UpdatedAt:         timestampToString(s.UpdatedAt),
	}
}

func (h *Handler) ListLeadSources(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	items, err := h.Queries.ListLeadSources(r.Context(), parseUUID(workspaceID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list lead sources")
		return
	}
	resp := make([]LeadSourceResponse, 0, len(items))
	for _, s := range items {
		resp = append(resp, leadSourceToResponse(s))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetLeadSource(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	s, err := h.Queries.GetLeadSource(r.Context(), parseUUID(id))
	if err != nil {
		writeError(w, http.StatusNotFound, "lead source not found")
		return
	}
	if uuidToString(s.WorkspaceID) != workspaceID {
		writeError(w, http.StatusForbidden, "lead source does not belong to workspace")
		return
	}
	writeJSON(w, http.StatusOK, leadSourceToResponse(s))
}

type CreateLeadSourceRequest struct {
	Name              string         `json:"name"`
	Slug              string         `json:"slug"`
	Provider          string         `json:"provider"`
	Config            map[string]any `json:"config"`
	IsActive          *bool          `json:"is_active"`
	AutoApprove       *bool          `json:"auto_approve"`
	EnrichmentEnabled *bool          `json:"enrichment_enabled"`
}

var validLeadSourceProviders = map[string]bool{
	"manual": true, "csv": true, "api": true, "form": true,
	"apollo": true, "hunter": true, "linkedin": true, "referral": true,
	"website": true, "hubspot": true, "pipedrive": true,
}

func (h *Handler) CreateLeadSource(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	var req CreateLeadSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "name and slug are required")
		return
	}
	if !validLeadSourceProviders[req.Provider] {
		req.Provider = "manual"
	}
	config := []byte("{}")
	if req.Config != nil {
		b, _ := json.Marshal(req.Config)
		if len(b) > 0 {
			config = b
		}
	}
	s, err := h.Queries.CreateLeadSource(r.Context(), db.CreateLeadSourceParams{
		WorkspaceID:       parseUUID(workspaceID),
		Name:              req.Name,
		Slug:              req.Slug,
		Provider:          req.Provider,
		Config:            config,
		IsActive:          pgtype.Bool{Bool: req.IsActive != nil && *req.IsActive, Valid: true},
		AutoApprove:       pgtype.Bool{Bool: req.AutoApprove != nil && *req.AutoApprove, Valid: true},
		EnrichmentEnabled: pgtype.Bool{Bool: req.EnrichmentEnabled == nil || *req.EnrichmentEnabled, Valid: true},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create lead source")
		return
	}
	writeJSON(w, http.StatusCreated, leadSourceToResponse(s))
}

type UpdateLeadSourceRequest struct {
	Name              *string        `json:"name"`
	Slug              *string        `json:"slug"`
	Provider          *string        `json:"provider"`
	Config            map[string]any `json:"config"`
	IsActive          *bool          `json:"is_active"`
	AutoApprove       *bool          `json:"auto_approve"`
	EnrichmentEnabled *bool          `json:"enrichment_enabled"`
}

func (h *Handler) UpdateLeadSource(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	var req UpdateLeadSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	var config []byte
	if req.Config != nil {
		b, _ := json.Marshal(req.Config)
		if len(b) > 0 {
			config = b
		}
	}
	params := db.UpdateLeadSourceParams{
		ID:        parseUUID(id),
		WorkspaceID: parseUUID(workspaceID),
	}
	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}
	if req.Slug != nil {
		params.Slug = pgtype.Text{String: *req.Slug, Valid: true}
	}
	if req.Provider != nil && validLeadSourceProviders[*req.Provider] {
		params.Provider = pgtype.Text{String: *req.Provider, Valid: true}
	}
	if len(config) > 0 {
		params.Config = config
	}
	if req.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *req.IsActive, Valid: true}
	}
	if req.AutoApprove != nil {
		params.AutoApprove = pgtype.Bool{Bool: *req.AutoApprove, Valid: true}
	}
	if req.EnrichmentEnabled != nil {
		params.EnrichmentEnabled = pgtype.Bool{Bool: *req.EnrichmentEnabled, Valid: true}
	}

	s, err := h.Queries.UpdateLeadSource(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update lead source")
		return
	}
	writeJSON(w, http.StatusOK, leadSourceToResponse(s))
}

func (h *Handler) DeleteLeadSource(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.Queries.DeleteLeadSource(r.Context(), db.DeleteLeadSourceParams{
		ID:        parseUUID(id),
		WorkspaceID: parseUUID(workspaceID),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete lead source")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// LeadImportBatch
// ---------------------------------------------------------------------------

type LeadImportBatchResponse struct {
	ID              string         `json:"id"`
	WorkspaceID     string         `json:"workspace_id"`
	SourceID        *string        `json:"source_id"`
	FileName        *string        `json:"file_name"`
	Provider        string         `json:"provider"`
	TotalRows       int32          `json:"total_rows"`
	ImportedCount   int32          `json:"imported_count"`
	DuplicateCount  int32          `json:"duplicate_count"`
	RejectedCount   int32          `json:"rejected_count"`
	Status          string         `json:"status"`
	ErrorLog        *string        `json:"error_log"`
	Metadata        map[string]any `json:"metadata"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
}

func leadImportBatchToResponse(b db.LeadImportBatch) LeadImportBatchResponse {
	meta := make(map[string]any)
	if len(b.Metadata) > 0 {
		_ = json.Unmarshal(b.Metadata, &meta)
	}
	resp := LeadImportBatchResponse{
		ID:             uuidToString(b.ID),
		WorkspaceID:    uuidToString(b.WorkspaceID),
		Provider:       b.Provider,
		TotalRows:      b.TotalRows,
		ImportedCount:  b.ImportedCount,
		DuplicateCount: b.DuplicateCount,
		RejectedCount:  b.RejectedCount,
		Status:         b.Status,
		Metadata:       meta,
		CreatedAt:      timestampToString(b.CreatedAt),
		UpdatedAt:      timestampToString(b.UpdatedAt),
	}
	if b.SourceID.Valid {
		s := uuidToString(b.SourceID)
		resp.SourceID = &s
	}
	if b.FileName.Valid {
		resp.FileName = &b.FileName.String
	}
	if b.ErrorLog.Valid {
		resp.ErrorLog = &b.ErrorLog.String
	}
	return resp
}

func (h *Handler) ListLeadImportBatches(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	limit := int32(50)
	offset := int32(0)
	items, err := h.Queries.ListLeadImportBatches(r.Context(), db.ListLeadImportBatchesParams{
		WorkspaceID: parseUUID(workspaceID),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list import batches")
		return
	}
	resp := make([]LeadImportBatchResponse, 0, len(items))
	for _, b := range items {
		resp = append(resp, leadImportBatchToResponse(b))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetLeadImportBatch(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	b, err := h.Queries.GetLeadImportBatch(r.Context(), parseUUID(id))
	if err != nil {
		writeError(w, http.StatusNotFound, "import batch not found")
		return
	}
	if uuidToString(b.WorkspaceID) != workspaceID {
		writeError(w, http.StatusForbidden, "import batch does not belong to workspace")
		return
	}
	writeJSON(w, http.StatusOK, leadImportBatchToResponse(b))
}

// ---------------------------------------------------------------------------
// LeadCuratorRule
// ---------------------------------------------------------------------------

type LeadCuratorRuleResponse struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Action      string `json:"action"`
	Field       string `json:"field"`
	Operator    string `json:"operator"`
	Value       *string `json:"value"`
	Priority    int32  `json:"priority"`
	IsActive    bool   `json:"is_active"`
	MatchCount  int32  `json:"match_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func leadCuratorRuleToResponse(rule db.LeadCuratorRule) LeadCuratorRuleResponse {
	resp := LeadCuratorRuleResponse{
		ID:          uuidToString(rule.ID),
		WorkspaceID: uuidToString(rule.WorkspaceID),
		Name:        rule.Name,
		Action:      rule.Action,
		Field:       rule.Field,
		Operator:    rule.Operator,
		Priority:    rule.Priority,
		IsActive:    rule.IsActive,
		MatchCount:  rule.MatchCount,
		CreatedAt:   timestampToString(rule.CreatedAt),
		UpdatedAt:   timestampToString(rule.UpdatedAt),
	}
	if rule.Value.Valid {
		resp.Value = &rule.Value.String
	}
	return resp
}

func (h *Handler) ListLeadCuratorRules(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	rules, err := h.Queries.ListLeadCuratorRules(r.Context(), parseUUID(workspaceID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list curator rules")
		return
	}
	resp := make([]LeadCuratorRuleResponse, 0, len(rules))
	for _, rule := range rules {
		resp = append(resp, leadCuratorRuleToResponse(rule))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetLeadCuratorRule(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	rule, err := h.Queries.GetLeadCuratorRule(r.Context(), parseUUID(id))
	if err != nil {
		writeError(w, http.StatusNotFound, "curator rule not found")
		return
	}
	if uuidToString(rule.WorkspaceID) != workspaceID {
		writeError(w, http.StatusForbidden, "curator rule does not belong to workspace")
		return
	}
	writeJSON(w, http.StatusOK, leadCuratorRuleToResponse(rule))
}

type CreateLeadCuratorRuleRequest struct {
	Name     string `json:"name"`
	Action   string `json:"action"`
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value,omitempty"`
	Priority int32  `json:"priority"`
	IsActive *bool  `json:"is_active"`
}

var validCuratorActions = map[string]bool{"approve": true, "reject": true, "review": true}
var validCuratorFields = map[string]bool{
	"email": true, "company": true, "name": true, "title": true,
	"industry": true, "company_size": true, "icp_fit": true,
	"budget": true, "authority": true, "need": true, "timeline": true,
}
var validCuratorOperators = map[string]bool{
	"exists": true, "not_exists": true, "contains": true, "not_contains": true,
	"eq": true, "ne": true, "gt": true, "gte": true, "lt": true, "lte": true,
	"regex": true, "domain_in": true, "domain_not_in": true,
}

func (h *Handler) CreateLeadCuratorRule(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	var req CreateLeadCuratorRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if !validCuratorActions[req.Action] {
		req.Action = "review"
	}
	if !validCuratorFields[req.Field] {
		writeError(w, http.StatusBadRequest, "invalid field")
		return
	}
	if !validCuratorOperators[req.Operator] {
		writeError(w, http.StatusBadRequest, "invalid operator")
		return
	}

	rule, err := h.Queries.CreateLeadCuratorRule(r.Context(), db.CreateLeadCuratorRuleParams{
		WorkspaceID: parseUUID(workspaceID),
		Name:        req.Name,
		Action:      req.Action,
		Field:       req.Field,
		Operator:    req.Operator,
		Value:       pgtype.Text{String: req.Value, Valid: req.Value != ""},
		Priority:    pgtype.Int4{Int32: req.Priority, Valid: true},
		IsActive:    pgtype.Bool{Bool: req.IsActive == nil || *req.IsActive, Valid: true},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create curator rule")
		return
	}
	writeJSON(w, http.StatusCreated, leadCuratorRuleToResponse(rule))
}

type UpdateLeadCuratorRuleRequest struct {
	Name     *string `json:"name"`
	Action   *string `json:"action"`
	Field    *string `json:"field"`
	Operator *string `json:"operator"`
	Value    *string `json:"value"`
	Priority *int32  `json:"priority"`
	IsActive *bool   `json:"is_active"`
}

func (h *Handler) UpdateLeadCuratorRule(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	var req UpdateLeadCuratorRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params := db.UpdateLeadCuratorRuleParams{
		ID:          parseUUID(id),
		WorkspaceID: parseUUID(workspaceID),
	}
	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}
	if req.Action != nil && validCuratorActions[*req.Action] {
		params.Action = pgtype.Text{String: *req.Action, Valid: true}
	}
	if req.Field != nil && validCuratorFields[*req.Field] {
		params.Field = pgtype.Text{String: *req.Field, Valid: true}
	}
	if req.Operator != nil && validCuratorOperators[*req.Operator] {
		params.Operator = pgtype.Text{String: *req.Operator, Valid: true}
	}
	if req.Value != nil {
		params.Value = pgtype.Text{String: *req.Value, Valid: true}
	}
	if req.Priority != nil {
		params.Priority = pgtype.Int4{Int32: *req.Priority, Valid: true}
	}
	if req.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *req.IsActive, Valid: true}
	}

	rule, err := h.Queries.UpdateLeadCuratorRule(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update curator rule")
		return
	}
	writeJSON(w, http.StatusOK, leadCuratorRuleToResponse(rule))
}

func (h *Handler) DeleteLeadCuratorRule(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.Queries.DeleteLeadCuratorRule(r.Context(), db.DeleteLeadCuratorRuleParams{
		ID:          parseUUID(id),
		WorkspaceID: parseUUID(workspaceID),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete curator rule")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Lead curation actions (approve / reject / bulk)
// ---------------------------------------------------------------------------

type CurationActionRequest struct {
	LeadIDs []string `json:"lead_ids"`
	Action  string   `json:"action"` // "approve" or "reject"
	Reason  string   `json:"reason,omitempty"`
}

func (h *Handler) BulkCurateLeads(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	member, ok := h.workspaceMember(w, r, workspaceID)
	if !ok {
		return
	}
	var req CurationActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.LeadIDs) == 0 {
		writeError(w, http.StatusBadRequest, "lead_ids is required")
		return
	}
	if req.Action != "approve" && req.Action != "reject" {
		writeError(w, http.StatusBadRequest, "action must be approve or reject")
		return
	}

	ctx := r.Context()
	wsID := parseUUID(workspaceID)
	memberID := member.ID

	approved := 0
	rejected := 0
	for _, leadIDStr := range req.LeadIDs {
		leadID := parseUUID(leadIDStr)
		lead, err := h.Queries.GetLeadInWorkspace(ctx, db.GetLeadInWorkspaceParams{
			ID:          leadID,
			WorkspaceID: wsID,
		})
		if err != nil {
			continue
		}

		newStatus := "captured"
		if req.Action == "approve" {
			newStatus = "captured"
			approved++
		} else {
			newStatus = "rejected"
			rejected++
		}

		stateMachine := lead.StateMachineStatus
		if stateMachine == "" || stateMachine == "captured" {
			stateMachine = newStatus
		}

		_, err = h.Queries.UpdateLead(ctx, db.UpdateLeadParams{
			ID:                 leadID,
			WorkspaceID:        wsID,
			Status:             pgtype.Text{String: newStatus, Valid: true},
			StateMachineStatus: pgtype.Text{String: stateMachine, Valid: true},
			LastEvent:          pgtype.Text{String: "curated." + req.Action, Valid: true},
		})
		if err != nil {
			continue
		}

		// Track who curated
		_, _ = h.DB.Exec(ctx,
			"UPDATE lead SET curated_at = now(), curated_by = $1 WHERE id = $2 AND workspace_id = $3",
			memberID, leadID, wsID)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"approved": approved,
		"rejected": rejected,
		"action":   req.Action,
	})
}

// EvaluateCuratorRules runs the active rules against a lead and returns the recommended action.
func (h *Handler) EvaluateCuratorRules(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	lead, err := h.Queries.GetLeadInWorkspace(r.Context(), db.GetLeadInWorkspaceParams{
		ID:          parseUUID(id),
		WorkspaceID: parseUUID(workspaceID),
	})
	if err != nil {
		writeError(w, http.StatusNotFound, "lead not found")
		return
	}

	rules, err := h.Queries.ListLeadCuratorRules(r.Context(), parseUUID(workspaceID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load rules")
		return
	}

	type matchResult struct {
		RuleID   string `json:"rule_id"`
		RuleName string `json:"rule_name"`
		Action   string `json:"action"`
		Reason   string `json:"reason"`
	}
	var matches []matchResult
	finalAction := "review"

	for _, rule := range rules {
		if !rule.IsActive {
			continue
		}
		matched, reason := evaluateRule(rule, lead)
		if matched {
			matches = append(matches, matchResult{
				RuleID:   uuidToString(rule.ID),
				RuleName: rule.Name,
				Action:   rule.Action,
				Reason:   reason,
			})
			if rule.Action == "reject" {
				finalAction = "reject"
			} else if rule.Action == "approve" && finalAction != "reject" {
				finalAction = "approve"
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"lead_id":      id,
		"recommended":  finalAction,
		"matched_rules": matches,
	})
}

func evaluateRule(rule db.LeadCuratorRule, lead db.Lead) (bool, string) {
	var fieldValue string
	switch rule.Field {
	case "email":
		fieldValue = lead.Email
	case "company":
		fieldValue = lead.Company
	case "name":
		fieldValue = lead.Name
	case "title":
		fieldValue = lead.Title
	case "industry":
		fieldValue = lead.Industry
	case "company_size":
		fieldValue = lead.CompanySize
	case "icp_fit":
		fieldValue = lead.IcpFit
	case "budget":
		fieldValue = lead.Budget
	case "authority":
		fieldValue = lead.Authority
	case "need":
		fieldValue = lead.Need
	case "timeline":
		fieldValue = lead.Timeline
	default:
		return false, ""
	}

	val := ""
	if rule.Value.Valid {
		val = rule.Value.String
	}

	switch rule.Operator {
	case "exists":
		return fieldValue != "" && fieldValue != "unknown", "field has value"
	case "not_exists":
		return fieldValue == "" || fieldValue == "unknown", "field is empty or unknown"
	case "contains":
		return strings.Contains(strings.ToLower(fieldValue), strings.ToLower(val)), "contains " + val
	case "not_contains":
		return !strings.Contains(strings.ToLower(fieldValue), strings.ToLower(val)), "does not contain " + val
	case "eq":
		return fieldValue == val, "equals " + val
	case "ne":
		return fieldValue != val, "not equals " + val
	case "domain_in":
		if rule.Field != "email" {
			return false, ""
		}
		parts := strings.Split(fieldValue, "@")
		if len(parts) != 2 {
			return false, ""
		}
		domain := parts[1]
		domains := strings.Split(val, ",")
		for _, d := range domains {
			if strings.TrimSpace(strings.ToLower(d)) == strings.ToLower(domain) {
				return true, "domain in blacklist"
			}
		}
		return false, ""
	case "domain_not_in":
		if rule.Field != "email" {
			return false, ""
		}
		parts := strings.Split(fieldValue, "@")
		if len(parts) != 2 {
			return false, ""
		}
		domain := parts[1]
		domains := strings.Split(val, ",")
		for _, d := range domains {
			if strings.TrimSpace(strings.ToLower(d)) == strings.ToLower(domain) {
				return false, ""
			}
		}
		return true, "domain not in whitelist"
	}
	return false, ""
}
