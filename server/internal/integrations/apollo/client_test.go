package apollo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchPeopleUsesAPIKeyHeaderAndFilters(t *testing.T) {
	var gotPath string
	var gotAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAPIKey = r.Header.Get("X-Api-Key")
		if r.URL.Query().Get("person_titles[]") != "Founder" {
			t.Fatalf("missing person_titles filter: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"people": []map[string]any{{
				"id":    "person-1",
				"name":  "Ana Silva",
				"title": "Founder",
				"organization": map[string]any{
					"name":        "AI Brasil",
					"website_url": "https://aibrasil.example",
				},
			}},
			"pagination": map[string]any{"page": 1, "per_page": 10, "total_entries": 1, "total_pages": 1},
		})
	}))
	defer srv.Close()

	client := NewClient(Config{APIKey: "test-key", BaseURL: srv.URL}, srv.Client())
	resp, err := client.SearchPeople(t.Context(), SearchRequest{
		PersonTitles:          []string{"Founder"},
		OrganizationLocations: []string{"Brazil"},
		Page:                  1,
		PerPage:               10,
	})
	if err != nil {
		t.Fatalf("SearchPeople returned error: %v", err)
	}
	if gotPath != "/api/v1/mixed_people/api_search" {
		t.Fatalf("path=%s", gotPath)
	}
	if gotAPIKey != "test-key" {
		t.Fatalf("X-Api-Key header=%q", gotAPIKey)
	}
	if len(resp.People) != 1 || resp.People[0].ID != "person-1" {
		t.Fatalf("unexpected response: %#v", resp.People)
	}
}

func TestBulkEnrichDisablesPersonalPhoneAndWaterfall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		for _, key := range []string{"reveal_personal_emails", "reveal_phone_number", "run_waterfall_email", "run_waterfall_phone"} {
			if q.Get(key) != "false" {
				t.Fatalf("%s=%q, want false", key, q.Get(key))
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"people": []map[string]any{{
				"id":           "person-1",
				"email":        "ana@aibrasil.example",
				"email_status": "verified",
			}},
		})
	}))
	defer srv.Close()

	client := NewClient(Config{APIKey: "test-key", BaseURL: srv.URL}, srv.Client())
	resp, err := client.BulkEnrichPeople(t.Context(), BulkEnrichRequest{
		Details: []EnrichPerson{{ID: "person-1", Name: "Ana Silva", Domain: "aibrasil.example"}},
	})
	if err != nil {
		t.Fatalf("BulkEnrichPeople returned error: %v", err)
	}
	if len(resp.People) != 1 || resp.People[0].Email != "ana@aibrasil.example" {
		t.Fatalf("unexpected enrichment response: %#v", resp.People)
	}
}

func TestBulkEnrichRejectsMoreThanTenPeople(t *testing.T) {
	client := NewClient(Config{APIKey: "test-key", BaseURL: "https://example.invalid"}, nil)
	req := BulkEnrichRequest{Details: make([]EnrichPerson, 11)}
	if _, err := client.BulkEnrichPeople(t.Context(), req); err == nil {
		t.Fatal("expected error for more than 10 people")
	}
}
