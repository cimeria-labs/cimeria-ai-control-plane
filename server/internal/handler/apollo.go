package handler

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/multica-ai/multica/server/internal/integrations/apollo"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
	"github.com/multica-ai/multica/server/pkg/protocol"
)

type ApolloStatusResponse struct {
	Configured bool `json:"configured"`
}

type ApolloSearchPreviewRequest struct {
	Titles                []string `json:"titles"`
	PersonLocations       []string `json:"person_locations"`
	OrganizationLocations []string `json:"organization_locations"`
	OrganizationKeywords  []string `json:"organization_keywords"`
	Seniorities           []string `json:"seniorities"`
	Limit                 int32    `json:"limit"`
}

type ApolloCandidateResponse struct {
	ID          string         `json:"id"`
	BatchID     string         `json:"batch_id"`
	ExternalID  string         `json:"external_id"`
	Email       *string        `json:"email"`
	EmailStatus *string        `json:"email_status"`
	Name        string         `json:"name"`
	Company     string         `json:"company"`
	Title       string         `json:"title"`
	Domain      string         `json:"domain"`
	LinkedInURL string         `json:"linkedin_url"`
	Status      string         `json:"status"`
	Score       int32          `json:"score"`
	Payload     map[string]any `json:"payload"`
}

type ApolloSearchPreviewResponse struct {
	BatchID    string                    `json:"batch_id"`
	Candidates []ApolloCandidateResponse `json:"candidates"`
}

type ApolloCandidateActionRequest struct {
	BatchID      string   `json:"batch_id"`
	CandidateIDs []string `json:"candidate_ids"`
}

type ApolloEnrichResponse struct {
	BatchID    string                    `json:"batch_id"`
	Candidates []ApolloCandidateResponse `json:"candidates"`
}

type ApolloImportRequest struct {
	BatchID      string   `json:"batch_id"`
	CandidateIDs []string `json:"candidate_ids"`
	NoSend       bool     `json:"no_send"`
}

type ApolloImportResponse struct {
	BatchID      string         `json:"batch_id"`
	Imported     int            `json:"imported"`
	Skipped      int            `json:"skipped"`
	MissingEmail int            `json:"missing_email"`
	Duplicates   int            `json:"duplicates"`
	Leads        []LeadResponse `json:"leads"`
}

func (h *Handler) ApolloStatus(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	writeJSON(w, http.StatusOK, ApolloStatusResponse{Configured: h.Apollo != nil && h.Apollo.Configured()})
}

func (h *Handler) ApolloSearchPreview(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	if h.Apollo == nil || !h.Apollo.Configured() {
		writeError(w, http.StatusServiceUnavailable, "apollo is not configured")
		return
	}

	var req ApolloSearchPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	limit := req.Limit
	if limit <= 0 || limit > 10 {
		limit = 10
	}

	source, err := h.ensureApolloLeadSource(r, workspaceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to prepare Apollo lead source")
		return
	}

	batchMetadata, _ := json.Marshal(map[string]any{
		"no_send":                true,
		"titles":                 req.Titles,
		"person_locations":       req.PersonLocations,
		"organization_locations": req.OrganizationLocations,
		"organization_keywords":  req.OrganizationKeywords,
		"seniorities":            req.Seniorities,
	})
	batch, err := h.Queries.CreateLeadImportBatch(r.Context(), db.CreateLeadImportBatchParams{
		WorkspaceID: parseUUID(workspaceID),
		Provider:    "apollo",
		TotalRows:   limit,
		SourceID:    source.ID,
		Status:      "preview",
		Metadata:    batchMetadata,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create Apollo preview batch")
		return
	}

	search, err := h.Apollo.SearchPeople(r.Context(), apollo.SearchRequest{
		PersonTitles:          req.Titles,
		PersonLocations:       req.PersonLocations,
		OrganizationLocations: req.OrganizationLocations,
		OrganizationKeywords:  req.OrganizationKeywords,
		PersonSeniorities:     req.Seniorities,
		Page:                  1,
		PerPage:               int(limit),
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, "apollo search failed")
		return
	}

	candidates := make([]ApolloCandidateResponse, 0, len(search.People))
	for _, person := range search.People {
		candidate, err := h.upsertApolloCandidate(r, workspaceID, batch.ID, person, "preview")
		if err != nil {
			continue
		}
		candidates = append(candidates, apolloCandidateToResponse(candidate))
	}

	writeJSON(w, http.StatusOK, ApolloSearchPreviewResponse{
		BatchID:    uuidToString(batch.ID),
		Candidates: candidates,
	})
}

func (h *Handler) ApolloEnrichCandidates(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}
	if h.Apollo == nil || !h.Apollo.Configured() {
		writeError(w, http.StatusServiceUnavailable, "apollo is not configured")
		return
	}

	var req ApolloCandidateActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	batchID := parseUUID(req.BatchID)
	if !batchID.Valid {
		writeError(w, http.StatusBadRequest, "valid batch_id is required")
		return
	}
	candidateIDs, ok := parseUUIDSlice(req.CandidateIDs)
	if !ok || len(candidateIDs) == 0 || len(candidateIDs) > 10 {
		writeError(w, http.StatusBadRequest, "candidate_ids must contain 1 to 10 valid ids")
		return
	}

	candidates, err := h.Queries.ListLeadImportCandidatesByIDs(r.Context(), db.ListLeadImportCandidatesByIDsParams{
		WorkspaceID: parseUUID(workspaceID),
		Ids:         candidateIDs,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load candidates")
		return
	}

	details := make([]apollo.EnrichPerson, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.BatchID != batchID || candidate.Status == "imported" {
			continue
		}
		details = append(details, apollo.EnrichPerson{
			ID:          candidate.ExternalID,
			Name:        candidate.Name,
			Domain:      candidate.Domain,
			CompanyName: candidate.Company,
			LinkedInURL: candidate.LinkedinUrl,
		})
	}
	if len(details) == 0 {
		writeError(w, http.StatusBadRequest, "no valid candidates to enrich")
		return
	}

	enriched, err := h.Apollo.BulkEnrichPeople(r.Context(), apollo.BulkEnrichRequest{Details: details})
	if err != nil {
		writeError(w, http.StatusBadGateway, "apollo enrichment failed")
		return
	}

	updated := make([]ApolloCandidateResponse, 0, len(enriched.People))
	for i, person := range enriched.People {
		if strings.TrimSpace(person.ID) == "" && i < len(details) {
			person.ID = details[i].ID
		}
		candidate, err := h.upsertApolloCandidate(r, workspaceID, batchID, person, "enriched")
		if err != nil {
			continue
		}
		updated = append(updated, apolloCandidateToResponse(candidate))
	}

	writeJSON(w, http.StatusOK, ApolloEnrichResponse{BatchID: req.BatchID, Candidates: updated})
}

func (h *Handler) ApolloImportApproved(w http.ResponseWriter, r *http.Request) {
	workspaceID := h.resolveWorkspaceID(r)
	if _, ok := h.workspaceMember(w, r, workspaceID); !ok {
		return
	}

	var req ApolloImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !req.NoSend {
		writeError(w, http.StatusBadRequest, "no_send must be true for Apollo imports")
		return
	}
	batchID := parseUUID(req.BatchID)
	if !batchID.Valid {
		writeError(w, http.StatusBadRequest, "valid batch_id is required")
		return
	}
	candidateIDs, ok := parseUUIDSlice(req.CandidateIDs)
	if !ok || len(candidateIDs) == 0 {
		writeError(w, http.StatusBadRequest, "candidate_ids is required")
		return
	}

	candidates, err := h.Queries.ListLeadImportCandidatesByIDs(r.Context(), db.ListLeadImportCandidatesByIDsParams{
		WorkspaceID: parseUUID(workspaceID),
		Ids:         candidateIDs,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load candidates")
		return
	}

	leads := []LeadResponse{}
	imported := 0
	skipped := 0
	missingEmail := 0
	duplicates := 0

	for _, candidate := range candidates {
		if candidate.BatchID != batchID || candidate.Status == "imported" {
			skipped++
			continue
		}
		email := strings.TrimSpace(candidate.Email.String)
		if email == "" || !candidate.Email.Valid {
			missingEmail++
			_, _ = h.Queries.UpdateLeadImportCandidateStatus(r.Context(), db.UpdateLeadImportCandidateStatusParams{
				ID:          candidate.ID,
				WorkspaceID: candidate.WorkspaceID,
				Status:      "missing_email",
				Error:       pgtype.Text{String: "Apollo enrichment did not return a usable business email", Valid: true},
			})
			continue
		}
		normalized, ok := normalizeLeadEmail(email)
		if !ok {
			skipped++
			_, _ = h.Queries.UpdateLeadImportCandidateStatus(r.Context(), db.UpdateLeadImportCandidateStatusParams{
				ID:          candidate.ID,
				WorkspaceID: candidate.WorkspaceID,
				Status:      "failed",
				Error:       pgtype.Text{String: "invalid email returned by Apollo", Valid: true},
			})
			continue
		}
		if h.leadExistsByEmail(r.Context(), workspaceID, normalized) {
			duplicates++
			_, _ = h.Queries.UpdateLeadImportCandidateStatus(r.Context(), db.UpdateLeadImportCandidateStatusParams{
				ID:          candidate.ID,
				WorkspaceID: candidate.WorkspaceID,
				Status:      "duplicate",
				Error:       pgtype.Text{String: "lead already exists", Valid: true},
			})
			continue
		}

		metadata := map[string]any{
			"source_provider":     "apollo",
			"apollo_person_id":    candidate.ExternalID,
			"apollo_no_send":      true,
			"apollo_email_status": candidate.EmailStatus.String,
			"import_batch_id":     req.BatchID,
		}
		metadataBytes, _ := json.Marshal(metadata)

		lead, err := h.Queries.CreateLead(r.Context(), db.CreateLeadParams{
			WorkspaceID:        parseUUID(workspaceID),
			Lower:              normalized,
			Name:               candidate.Name,
			Company:            candidate.Company,
			Title:              candidate.Title,
			Source:             "apollo",
			Status:             "captured",
			StateMachineStatus: "captured",
			LastEvent:          pgtype.Text{String: "apollo.imported_no_send", Valid: true},
			Metadata:           metadataBytes,
			IcpFit:             "unknown",
			LeadTemperature:    "cold",
		})
		if err != nil {
			skipped++
			_, _ = h.Queries.UpdateLeadImportCandidateStatus(r.Context(), db.UpdateLeadImportCandidateStatusParams{
				ID:          candidate.ID,
				WorkspaceID: candidate.WorkspaceID,
				Status:      "failed",
				Error:       pgtype.Text{String: "lead could not be created", Valid: true},
			})
			continue
		}

		if h.applyCuratorRules(r.Context(), lead) == "reject" {
			if updatedLead, err := h.Queries.UpdateLead(r.Context(), db.UpdateLeadParams{
				ID:                 lead.ID,
				WorkspaceID:        lead.WorkspaceID,
				Status:             pgtype.Text{String: "rejected", Valid: true},
				StateMachineStatus: pgtype.Text{String: "rejected", Valid: true},
				LastEvent:          pgtype.Text{String: "curator.auto_rejected", Valid: true},
			}); err == nil {
				lead = updatedLead
			}
		}

		_ = h.Queries.SetLeadImportBatch(r.Context(), db.SetLeadImportBatchParams{
			ImportBatchID: batchID,
			ID:            lead.ID,
			WorkspaceID:   parseUUID(workspaceID),
		})
		_, _ = h.Queries.MarkLeadImportCandidateImported(r.Context(), db.MarkLeadImportCandidateImportedParams{
			ID:          candidate.ID,
			WorkspaceID: candidate.WorkspaceID,
			LeadID:      lead.ID,
		})

		resp := leadToResponse(lead)
		h.publish(protocol.EventLeadCreated, workspaceID, "system", "", map[string]any{"lead": resp})
		leads = append(leads, resp)
		imported++
	}

	_, _ = h.Queries.UpdateLeadImportBatch(r.Context(), db.UpdateLeadImportBatchParams{
		ID:             batchID,
		ImportedCount:  pgtype.Int4{Int32: int32(imported), Valid: true},
		DuplicateCount: pgtype.Int4{Int32: int32(duplicates), Valid: true},
		RejectedCount:  pgtype.Int4{Int32: int32(skipped + missingEmail), Valid: true},
		Status:         pgtype.Text{String: "completed", Valid: true},
	})

	writeJSON(w, http.StatusOK, ApolloImportResponse{
		BatchID:      req.BatchID,
		Imported:     imported,
		Skipped:      skipped,
		MissingEmail: missingEmail,
		Duplicates:   duplicates,
		Leads:        leads,
	})
}

func (h *Handler) ensureApolloLeadSource(r *http.Request, workspaceID string) (db.LeadSource, error) {
	existing, err := h.Queries.GetLeadSourceBySlug(r.Context(), db.GetLeadSourceBySlugParams{
		WorkspaceID: parseUUID(workspaceID),
		Slug:        "apollo",
	})
	if err == nil {
		return existing, nil
	}
	return h.Queries.CreateLeadSource(r.Context(), db.CreateLeadSourceParams{
		WorkspaceID:       parseUUID(workspaceID),
		Name:              "Apollo",
		Slug:              "apollo",
		Provider:          "apollo",
		Config:            []byte(`{"managed_by":"cimeria","stores_secret":false}`),
		IsActive:          pgtype.Bool{Bool: true, Valid: true},
		AutoApprove:       pgtype.Bool{Bool: false, Valid: true},
		EnrichmentEnabled: pgtype.Bool{Bool: true, Valid: true},
	})
}

func (h *Handler) upsertApolloCandidate(r *http.Request, workspaceID string, batchID pgtype.UUID, person apollo.Person, status string) (db.LeadImportCandidate, error) {
	payload, _ := json.Marshal(person)
	name := strings.TrimSpace(person.Name)
	if name == "" {
		name = strings.TrimSpace(strings.TrimSpace(person.FirstName + " " + person.LastName))
	}
	domain := apolloOrganizationDomain(person.Organization)
	externalID := strings.TrimSpace(person.ID)
	if externalID == "" {
		externalID = apolloFallbackExternalID(name, domain, person.LinkedInURL)
	}
	return h.Queries.UpsertLeadImportCandidate(r.Context(), db.UpsertLeadImportCandidateParams{
		WorkspaceID: parseUUID(workspaceID),
		BatchID:     batchID,
		Provider:    "apollo",
		ExternalID:  externalID,
		Email:       pgtype.Text{String: strings.TrimSpace(person.Email), Valid: strings.TrimSpace(person.Email) != ""},
		EmailStatus: pgtype.Text{String: strings.TrimSpace(person.EmailStatus), Valid: strings.TrimSpace(person.EmailStatus) != ""},
		Name:        name,
		Company:     person.Organization.Name,
		Title:       person.Title,
		Domain:      domain,
		LinkedinUrl: person.LinkedInURL,
		Status:      status,
		Score:       int32(scoreApolloCandidate(person)),
		Payload:     payload,
	})
}

func apolloCandidateToResponse(c db.LeadImportCandidate) ApolloCandidateResponse {
	payload := map[string]any{}
	if len(c.Payload) > 0 {
		_ = json.Unmarshal(c.Payload, &payload)
	}
	return ApolloCandidateResponse{
		ID:          uuidToString(c.ID),
		BatchID:     uuidToString(c.BatchID),
		ExternalID:  c.ExternalID,
		Email:       textToPtr(c.Email),
		EmailStatus: textToPtr(c.EmailStatus),
		Name:        c.Name,
		Company:     c.Company,
		Title:       c.Title,
		Domain:      c.Domain,
		LinkedInURL: c.LinkedinUrl,
		Status:      c.Status,
		Score:       c.Score,
		Payload:     payload,
	}
}

func scoreApolloCandidate(person apollo.Person) int {
	score := 1
	if strings.TrimSpace(person.Title) != "" {
		score += 2
	}
	if strings.TrimSpace(person.Organization.Name) != "" {
		score += 2
	}
	if strings.TrimSpace(person.Organization.WebsiteURL) != "" || strings.TrimSpace(person.Organization.Domain) != "" {
		score += 2
	}
	if strings.TrimSpace(person.LinkedInURL) != "" {
		score++
	}
	if strings.TrimSpace(person.Email) != "" {
		score += 2
	}
	return score
}

func apolloOrganizationDomain(org apollo.Organization) string {
	domain := strings.TrimSpace(org.Domain)
	if domain == "" {
		domain = strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(org.WebsiteURL), "https://"), "http://")
		domain = strings.TrimPrefix(domain, "www.")
		domain = strings.Split(domain, "/")[0]
	}
	return strings.ToLower(strings.TrimSpace(domain))
}

func apolloFallbackExternalID(parts ...string) string {
	raw := strings.ToLower(strings.Join(parts, "|"))
	sum := sha1.Sum([]byte(raw))
	return "fallback-" + hex.EncodeToString(sum[:])
}

func parseUUIDSlice(ids []string) ([]pgtype.UUID, bool) {
	out := make([]pgtype.UUID, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		parsed := parseUUID(id)
		if !parsed.Valid {
			return nil, false
		}
		out = append(out, parsed)
	}
	return out, true
}

func (h *Handler) leadExistsByEmail(ctx context.Context, workspaceID, email string) bool {
	if h.DB == nil {
		return false
	}
	var id pgtype.UUID
	err := h.DB.QueryRow(ctx, `SELECT id FROM lead WHERE workspace_id = $1 AND email = lower($2)`, parseUUID(workspaceID), email).Scan(&id)
	return err == nil
}
