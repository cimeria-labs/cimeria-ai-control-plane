package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
	"github.com/multica-ai/multica/server/pkg/protocol"
)

type LeadResponse struct {
	ID                 string         `json:"id"`
	WorkspaceID        string         `json:"workspace_id"`
	Email              string         `json:"email"`
	Name               string         `json:"name"`
	Company            string         `json:"company"`
	Title              string         `json:"title"`
	Source             string         `json:"source"`
	Status             string         `json:"status"`
	Score              int32          `json:"score"`
	DynamicScore       int32          `json:"dynamic_score"`
	AssigneeType       *string        `json:"assignee_type"`
	AssigneeID         *string        `json:"assignee_id"`
	PipelineID         *string        `json:"pipeline_id"`
	StateMachineStatus string         `json:"state_machine_status"`
	LastEvent          *string        `json:"last_event"`
	Metadata           map[string]any `json:"metadata"`
	Budget             string         `json:"budget"`
	Authority          string         `json:"authority"`
	Need               string         `json:"need"`
	Timeline           string         `json:"timeline"`
	CompanySize        string         `json:"company_size"`
	Industry           string         `json:"industry"`
	PainPoints         string         `json:"pain_points"`
	IcpFit             string         `json:"icp_fit"`
	LeadTemperature    string         `json:"lead_temperature"`
	CreatedAt          string         `json:"created_at"`
	UpdatedAt          string         `json:"updated_at"`
}

type ListLeadsResponse struct {
	Leads []LeadResponse `json:"leads"`
	Total int64          `json:"total"`
}

type CreateLeadRequest struct {
	Email              string         `json:"email"`
	Name               *string        `json:"name"`
	Company            *string        `json:"company"`
	Title              *string        `json:"title"`
	Source             *string        `json:"source"`
	Status             *string        `json:"status"`
	Score              *int32         `json:"score"`
	DynamicScore       *int32         `json:"dynamic_score"`
	AssigneeType       *string        `json:"assignee_type"`
	AssigneeID         *string        `json:"assignee_id"`
	PipelineID         *string        `json:"pipeline_id"`
	StateMachineStatus *string        `json:"state_machine_status"`
	LastEvent          *string        `json:"last_event"`
	Metadata           map[string]any `json:"metadata"`
	Budget             *string        `json:"budget"`
	Authority          *string        `json:"authority"`
	Need               *string        `json:"need"`
	Timeline           *string        `json:"timeline"`
	CompanySize        *string        `json:"company_size"`
	Industry           *string        `json:"industry"`
	PainPoints         *string        `json:"pain_points"`
	IcpFit             *string        `json:"icp_fit"`
	LeadTemperature    *string        `json:"lead_temperature"`
}

type UpdateLeadRequest struct {
	Email              *string        `json:"email"`
	Name               *string        `json:"name"`
	Company            *string        `json:"company"`
	Title              *string        `json:"title"`
	Source             *string        `json:"source"`
	Status             *string        `json:"status"`
	Score              *int32         `json:"score"`
	DynamicScore       *int32         `json:"dynamic_score"`
	AssigneeType       *string        `json:"assignee_type"`
	AssigneeID         *string        `json:"assignee_id"`
	PipelineID         *string        `json:"pipeline_id"`
	StateMachineStatus *string        `json:"state_machine_status"`
	LastEvent          *string        `json:"last_event"`
	Metadata           map[string]any `json:"metadata"`
	Budget             *string        `json:"budget"`
	Authority          *string        `json:"authority"`
	Need               *string        `json:"need"`
	Timeline           *string        `json:"timeline"`
	CompanySize        *string        `json:"company_size"`
	Industry           *string        `json:"industry"`
	PainPoints         *string        `json:"pain_points"`
	IcpFit             *string        `json:"icp_fit"`
	LeadTemperature    *string        `json:"lead_temperature"`
}

type ImportLeadsResponse struct {
	Imported int            `json:"imported"`
	Skipped  int            `json:"skipped"`
	Leads    []LeadResponse `json:"leads"`
}

var validLeadStatuses = map[string]bool{
	"captured": true, "qualified": true, "rejected": true, "copy_ready": true,
	"strategy_ready": true, "email_sent": true, "nurturing": true, "hot": true,
	"handoff_human": true, "converted": true, "cancelled": true,
}

func leadToResponse(l db.Lead) LeadResponse {
	metadata := map[string]any{}
	if len(l.Metadata) > 0 {
		_ = json.Unmarshal(l.Metadata, &metadata)
	}
	return LeadResponse{
		ID:                 uuidToString(l.ID),
		WorkspaceID:        uuidToString(l.WorkspaceID),
		Email:              l.Email,
		Name:               l.Name,
		Company:            l.Company,
		Title:              l.Title,
		Source:             l.Source,
		Status:             l.Status,
		Score:              l.Score,
		DynamicScore:       l.DynamicScore,
		AssigneeType:       textToPtr(l.AssigneeType),
		AssigneeID:         uuidToPtr(l.AssigneeID),
		PipelineID:         uuidToPtr(l.PipelineID),
		StateMachineStatus: l.StateMachineStatus,
		LastEvent:          textToPtr(l.LastEvent),
		Metadata:           metadata,
		CreatedAt:          timestampToString(l.CreatedAt),
		UpdatedAt:          timestampToString(l.UpdatedAt),
	}
}

func normalizeLeadEmail(email string) (string, bool) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "", false
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", false
	}
	return email, true
}

func optionalString(v *string) any {
	if v == nil {
		return nil
	}
	return strings.TrimSpace(*v)
}

func optionalInt(v *int32) any {
	if v == nil {
		return nil
	}
	return *v
}

func optionalUUID(v *string) pgtype.UUID {
	if v == nil || strings.TrimSpace(*v) == "" {
		return pgtype.UUID{}
	}
	return parseUUID(strings.TrimSpace(*v))
}

func optionalText(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: strings.TrimSpace(*v), Valid: true}
}

func metadataBytes(metadata map[string]any) ([]byte, error) {
	if metadata == nil {
		return nil, nil
	}
	return json.Marshal(metadata)
}

func validateLeadStatus(status *string) bool {
	if status == nil {
		return true
	}
	return validLeadStatuses[strings.TrimSpace(*status)]
}

func (h *Handler) ListLeads(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}

	limit := int32(50)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 200 {
			limit = int32(parsed)
		}
	}
	offset := int32(0)
	if raw := r.URL.Query().Get("offset"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}
	status := pgtype.Text{}
	if raw := strings.TrimSpace(r.URL.Query().Get("status")); raw != "" {
		if !validLeadStatuses[raw] {
			writeError(w, http.StatusBadRequest, "invalid lead status")
			return
		}
		status = pgtype.Text{String: raw, Valid: true}
	}

	ctx := r.Context()
	params := db.ListLeadsParams{WorkspaceID: parseUUID(workspaceID), Limit: limit, Offset: offset, Status: status}
	leads, err := h.Queries.ListLeads(ctx, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list leads")
		return
	}
	total, err := h.Queries.CountLeads(ctx, db.CountLeadsParams{WorkspaceID: parseUUID(workspaceID), Status: status})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to count leads")
		return
	}
	resp := make([]LeadResponse, 0, len(leads))
	for _, lead := range leads {
		resp = append(resp, leadToResponse(lead))
	}
	writeJSON(w, http.StatusOK, ListLeadsResponse{Leads: resp, Total: total})
}

func (h *Handler) CreateLead(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	var req CreateLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	email, ok := normalizeLeadEmail(req.Email)
	if !ok {
		writeError(w, http.StatusBadRequest, "valid email is required")
		return
	}
	if !validateLeadStatus(req.Status) {
		writeError(w, http.StatusBadRequest, "invalid lead status")
		return
	}
	metadata, err := metadataBytes(req.Metadata)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid metadata")
		return
	}
	lead, err := h.Queries.CreateLead(r.Context(), db.CreateLeadParams{
		WorkspaceID:        parseUUID(workspaceID),
		Lower:              email,
		Name:               optionalString(req.Name),
		Company:            optionalString(req.Company),
		Title:              optionalString(req.Title),
		Source:             optionalString(req.Source),
		Status:             optionalString(req.Status),
		Score:              optionalInt(req.Score),
		DynamicScore:       optionalInt(req.DynamicScore),
		AssigneeType:       optionalText(req.AssigneeType),
		AssigneeID:         optionalUUID(req.AssigneeID),
		PipelineID:         optionalUUID(req.PipelineID),
		StateMachineStatus: optionalString(req.StateMachineStatus),
		LastEvent:          optionalText(req.LastEvent),
		Metadata:           metadata,
		Budget:             optionalText(req.Budget),
		Authority:          optionalText(req.Authority),
		Need:               optionalText(req.Need),
		Timeline:           optionalText(req.Timeline),
		CompanySize:        optionalText(req.CompanySize),
		Industry:           optionalText(req.Industry),
		PainPoints:         optionalText(req.PainPoints),
		IcpFit:             optionalText(req.IcpFit),
		LeadTemperature:    optionalText(req.LeadTemperature),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create lead")
		return
	}

	// Auto-curation: evaluate curator rules against the new lead
	curationAction := h.applyCuratorRules(r.Context(), lead)
	if curationAction == "reject" {
		_, _ = h.Queries.UpdateLead(r.Context(), db.UpdateLeadParams{
			ID:                 lead.ID,
			WorkspaceID:        lead.WorkspaceID,
			Status:             pgtype.Text{String: "rejected", Valid: true},
			StateMachineStatus: pgtype.Text{String: "rejected", Valid: true},
			LastEvent:          pgtype.Text{String: "curator.auto_rejected", Valid: true},
		})
		lead.Status = "rejected"
		lead.StateMachineStatus = "rejected"
		lead.LastEvent = pgtype.Text{String: "curator.auto_rejected", Valid: true}
	}

	resp := leadToResponse(lead)
	h.publish(protocol.EventLeadCreated, workspaceID, "system", "", map[string]any{"lead": resp})
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetLead(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	lead, err := h.Queries.GetLeadInWorkspace(r.Context(), db.GetLeadInWorkspaceParams{
		ID:          parseUUID(chi.URLParam(r, "id")),
		WorkspaceID: parseUUID(workspaceID),
	})
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "lead not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get lead")
		return
	}
	writeJSON(w, http.StatusOK, leadToResponse(lead))
}

func (h *Handler) UpdateLead(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	var req UpdateLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email != nil {
		email, ok := normalizeLeadEmail(*req.Email)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid email")
			return
		}
		req.Email = &email
	}
	if !validateLeadStatus(req.Status) {
		writeError(w, http.StatusBadRequest, "invalid lead status")
		return
	}
	metadata, err := metadataBytes(req.Metadata)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid metadata")
		return
	}
	lead, err := h.Queries.UpdateLead(r.Context(), db.UpdateLeadParams{
		ID:                 parseUUID(chi.URLParam(r, "id")),
		WorkspaceID:        parseUUID(workspaceID),
		Email:              optionalText(req.Email),
		Name:               optionalText(req.Name),
		Company:            optionalText(req.Company),
		Title:              optionalText(req.Title),
		Source:             optionalText(req.Source),
		Status:             optionalText(req.Status),
		Score:              pgtype.Int4{Int32: int32Value(req.Score), Valid: req.Score != nil},
		DynamicScore:       pgtype.Int4{Int32: int32Value(req.DynamicScore), Valid: req.DynamicScore != nil},
		AssigneeType:       optionalText(req.AssigneeType),
		AssigneeID:         optionalUUID(req.AssigneeID),
		PipelineID:         optionalUUID(req.PipelineID),
		StateMachineStatus: optionalText(req.StateMachineStatus),
		LastEvent:          optionalText(req.LastEvent),
		Metadata:           metadata,
		Budget:             optionalText(req.Budget),
		Authority:          optionalText(req.Authority),
		Need:               optionalText(req.Need),
		Timeline:           optionalText(req.Timeline),
		CompanySize:        optionalText(req.CompanySize),
		Industry:           optionalText(req.Industry),
		PainPoints:         optionalText(req.PainPoints),
		IcpFit:             optionalText(req.IcpFit),
		LeadTemperature:    optionalText(req.LeadTemperature),
	})
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "lead not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update lead")
		return
	}
	resp := leadToResponse(lead)
	switch lead.Status {
	case "qualified":
		h.publish(protocol.EventLeadQualified, workspaceID, "system", "", map[string]any{"lead": resp})
	case "rejected":
		h.publish(protocol.EventLeadRejected, workspaceID, "system", "", map[string]any{"lead": resp})
	case "handoff_human":
		h.publish(protocol.EventLeadHandoffHuman, workspaceID, "system", "", map[string]any{"lead": resp})
	}
	writeJSON(w, http.StatusOK, resp)
}

func int32Value(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

func (h *Handler) ImportLeadsCSV(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	reader := csv.NewReader(io.LimitReader(r.Body, 2<<20))
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid CSV")
		return
	}
	if len(rows) == 0 {
		writeJSON(w, http.StatusOK, ImportLeadsResponse{Leads: []LeadResponse{}})
		return
	}
	header := map[string]int{}
	for i, col := range rows[0] {
		header[strings.ToLower(strings.TrimSpace(col))] = i
	}
	value := func(row []string, name string) string {
		idx, ok := header[name]
		if !ok || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}
	imported := []LeadResponse{}
	skipped := 0
	for _, row := range rows[1:] {
		email, ok := normalizeLeadEmail(value(row, "email"))
		if !ok {
			skipped++
			continue
		}
		name, company, title := value(row, "name"), value(row, "company"), value(row, "title")
		source := value(row, "source")
		if source == "" {
			source = "csv"
		}
		lead, err := h.Queries.CreateLead(r.Context(), db.CreateLeadParams{
			WorkspaceID: parseUUID(workspaceID),
			Lower:       email,
			Name:        name,
			Company:     company,
			Title:       title,
			Source:      source,
			Status:      "captured",
		})
		if err != nil {
			skipped++
			continue
		}
		if h.applyCuratorRules(r.Context(), lead) == "reject" {
			_, _ = h.Queries.UpdateLead(r.Context(), db.UpdateLeadParams{
				ID:                 lead.ID,
				WorkspaceID:        lead.WorkspaceID,
				Status:             pgtype.Text{String: "rejected", Valid: true},
				StateMachineStatus: pgtype.Text{String: "rejected", Valid: true},
				LastEvent:          pgtype.Text{String: "curator.auto_rejected", Valid: true},
			})
			skipped++
			continue
		}
		imported = append(imported, leadToResponse(lead))
	}
	writeJSON(w, http.StatusOK, ImportLeadsResponse{Imported: len(imported), Skipped: skipped, Leads: imported})
}

// applyCuratorRules evaluates active curator rules against a lead.
// Returns the final action: "approve", "reject", or "review".
// Reject takes priority over approve. Increments match_count for matched rules.
func (h *Handler) applyCuratorRules(ctx context.Context, lead db.Lead) string {
	rules, err := h.Queries.ListLeadCuratorRules(ctx, lead.WorkspaceID)
	if err != nil {
		return "review"
	}

	finalAction := "review"
	for _, rule := range rules {
		if !rule.IsActive {
			continue
		}
		matched, _ := evaluateRule(rule, lead)
		if matched {
			_ = h.Queries.IncrementRuleMatchCount(ctx, rule.ID)
			if rule.Action == "reject" {
				finalAction = "reject"
			} else if rule.Action == "approve" && finalAction != "reject" {
				finalAction = "approve"
			}
		}
	}
	return finalAction
}