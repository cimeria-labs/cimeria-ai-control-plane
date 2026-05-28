package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

var validBatchProviders = map[string]bool{
	"manual":    true,
	"csv":       true,
	"api":       true,
	"form":      true,
	"apollo":    true,
	"hunter":    true,
	"linkedin":  true,
	"referral":  true,
	"website":   true,
	"hubspot":   true,
	"pipedrive": true,
}

type CreateLeadImportBatchRequest struct {
	Provider  string         `json:"provider"`
	TotalRows int32          `json:"total_rows"`
	FileName  *string        `json:"file_name,omitempty"`
	SourceID  *string        `json:"source_id,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

func (h *Handler) CreateLeadImportBatch(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	var req CreateLeadImportBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !validBatchProviders[req.Provider] {
		req.Provider = "api"
	}
	metadata := []byte("{}")
	if req.Metadata != nil {
		b, _ := json.Marshal(req.Metadata)
		if len(b) > 0 {
			metadata = b
		}
	}
	var sourceID pgtype.UUID
	if req.SourceID != nil {
		sourceID = parseUUID(*req.SourceID)
	}
	var fileName pgtype.Text
	if req.FileName != nil {
		fileName = pgtype.Text{String: *req.FileName, Valid: true}
	}
	b, err := h.Queries.CreateLeadImportBatch(r.Context(), db.CreateLeadImportBatchParams{
		WorkspaceID: parseUUID(workspaceID),
		Provider:    req.Provider,
		TotalRows:   req.TotalRows,
		SourceID:    sourceID,
		FileName:    fileName,
		Metadata:    metadata,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create batch")
		return
	}
	writeJSON(w, http.StatusCreated, leadImportBatchToResponse(b))
}

type UpdateLeadImportBatchRequest struct {
	ImportedCount  *int32  `json:"imported_count,omitempty"`
	DuplicateCount *int32  `json:"duplicate_count,omitempty"`
	RejectedCount  *int32  `json:"rejected_count,omitempty"`
	Status         *string `json:"status,omitempty"`
	ErrorLog       *string `json:"error_log,omitempty"`
}

func (h *Handler) UpdateLeadImportBatch(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	id := chi.URLParam(r, "id")
	b, err := h.Queries.GetLeadImportBatch(r.Context(), parseUUID(id))
	if err != nil {
		writeError(w, http.StatusNotFound, "batch not found")
		return
	}
	if uuidToString(b.WorkspaceID) != workspaceID {
		writeError(w, http.StatusForbidden, "batch does not belong to workspace")
		return
	}
	var req UpdateLeadImportBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, err := h.Queries.UpdateLeadImportBatch(r.Context(), db.UpdateLeadImportBatchParams{
		ID:             parseUUID(id),
		ImportedCount:  pgtype.Int4{Int32: int32Value(req.ImportedCount), Valid: req.ImportedCount != nil},
		DuplicateCount: pgtype.Int4{Int32: int32Value(req.DuplicateCount), Valid: req.DuplicateCount != nil},
		RejectedCount:  pgtype.Int4{Int32: int32Value(req.RejectedCount), Valid: req.RejectedCount != nil},
		Status:         pgtype.Text{String: stringValue(req.Status), Valid: req.Status != nil},
		ErrorLog:       pgtype.Text{String: stringValue(req.ErrorLog), Valid: req.ErrorLog != nil},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update batch")
		return
	}
	writeJSON(w, http.StatusOK, leadImportBatchToResponse(updated))
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
