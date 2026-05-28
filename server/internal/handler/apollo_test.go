package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/multica-ai/multica/server/internal/integrations/apollo"
)

type fakeApolloClient struct {
	configured bool
	searchResp apollo.SearchResponse
	enrichResp apollo.BulkEnrichResponse
}

func (f fakeApolloClient) Configured() bool { return f.configured }

func (f fakeApolloClient) SearchPeople(_ context.Context, _ apollo.SearchRequest) (apollo.SearchResponse, error) {
	return f.searchResp, nil
}

func (f fakeApolloClient) BulkEnrichPeople(_ context.Context, _ apollo.BulkEnrichRequest) (apollo.BulkEnrichResponse, error) {
	return f.enrichResp, nil
}

func (f fakeApolloClient) UsageStats(_ context.Context) (apollo.UsageStatsResponse, error) {
	return apollo.UsageStatsResponse{}, nil
}

func TestApolloStatusDoesNotExposeSecret(t *testing.T) {
	previous := testHandler.Apollo
	testHandler.Apollo = fakeApolloClient{configured: true}
	defer func() { testHandler.Apollo = previous }()

	w := httptest.NewRecorder()
	req := newRequest("GET", "/api/integrations/apollo/status", nil)
	testHandler.ApolloStatus(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ApolloStatus expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "APOLLO_API_KEY") || strings.Contains(w.Body.String(), "test-key") {
		t.Fatalf("status response leaked secret material: %s", w.Body.String())
	}
}

func TestApolloImportRequiresNoSend(t *testing.T) {
	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/integrations/apollo/import-approved", map[string]any{
		"batch_id":      testWorkspaceID,
		"candidate_ids": []string{testWorkspaceID},
		"no_send":       false,
	})
	testHandler.ApolloImportApproved(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ApolloImportApproved expected 400 without no_send, got %d: %s", w.Code, w.Body.String())
	}
}

func TestApolloPreviewEnrichAndImportNoSend(t *testing.T) {
	previous := testHandler.Apollo
	testHandler.Apollo = fakeApolloClient{
		configured: true,
		searchResp: apollo.SearchResponse{People: []apollo.Person{{
			ID:          "apollo-person-1",
			Name:        "Ana Silva",
			Title:       "Founder",
			LinkedInURL: "https://linkedin.example/ana",
			Organization: apollo.Organization{
				Name:       "AI Brasil",
				WebsiteURL: "https://aibrasil.example",
				Domain:     "aibrasil.example",
			},
		}}},
		enrichResp: apollo.BulkEnrichResponse{People: []apollo.Person{{
			ID:          "apollo-person-1",
			Name:        "Ana Silva",
			Title:       "Founder",
			Email:       "ana@aibrasil.example",
			EmailStatus: "verified",
			Organization: apollo.Organization{
				Name:       "AI Brasil",
				WebsiteURL: "https://aibrasil.example",
				Domain:     "aibrasil.example",
			},
		}}},
	}
	defer func() { testHandler.Apollo = previous }()

	w := httptest.NewRecorder()
	req := newRequest("POST", "/api/integrations/apollo/search-preview", map[string]any{
		"titles":                 []string{"Founder"},
		"organization_locations": []string{"Brazil"},
		"organization_keywords":  []string{"artificial intelligence"},
		"limit":                  1,
	})
	testHandler.ApolloSearchPreview(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ApolloSearchPreview expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var preview ApolloSearchPreviewResponse
	if err := json.NewDecoder(w.Body).Decode(&preview); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if len(preview.Candidates) != 1 || preview.Candidates[0].Email != nil {
		t.Fatalf("preview must create one candidate without email: %#v", preview.Candidates)
	}

	w = httptest.NewRecorder()
	req = newRequest("POST", "/api/integrations/apollo/enrich", map[string]any{
		"batch_id":      preview.BatchID,
		"candidate_ids": []string{preview.Candidates[0].ID},
	})
	testHandler.ApolloEnrichCandidates(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ApolloEnrichCandidates expected 200, got %d: %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	req = newRequest("POST", "/api/integrations/apollo/import-approved", map[string]any{
		"batch_id":      preview.BatchID,
		"candidate_ids": []string{preview.Candidates[0].ID},
		"no_send":       true,
	})
	testHandler.ApolloImportApproved(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ApolloImportApproved expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var imported ApolloImportResponse
	if err := json.NewDecoder(w.Body).Decode(&imported); err != nil {
		t.Fatalf("decode import: %v", err)
	}
	if imported.Imported != 1 || len(imported.Leads) != 1 {
		t.Fatalf("expected one imported lead: %#v", imported)
	}
	if imported.Leads[0].Source != "apollo" || imported.Leads[0].Status != "captured" {
		t.Fatalf("lead must be Apollo captured no-send: %#v", imported.Leads[0])
	}

	var emailLogCount int
	if err := testPool.QueryRow(t.Context(), `SELECT count(*) FROM email_log WHERE lead_id = $1`, parseUUID(imported.Leads[0].ID)).Scan(&emailLogCount); err != nil {
		t.Fatalf("count email logs: %v", err)
	}
	if emailLogCount != 0 {
		t.Fatalf("Apollo no-send import created email logs: %d", emailLogCount)
	}
}
