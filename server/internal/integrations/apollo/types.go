package apollo

import "encoding/json"

type Config struct {
	APIKey  string
	BaseURL string
}

func (c Config) EffectiveBaseURL() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return "https://api.apollo.io"
}

func (c Config) Configured() bool {
	return c.APIKey != ""
}

type SearchRequest struct {
	PersonTitles          []string `json:"person_titles,omitempty"`
	PersonLocations       []string `json:"person_locations,omitempty"`
	OrganizationLocations []string `json:"organization_locations,omitempty"`
	OrganizationKeywords  []string `json:"q_organization_keyword_tags,omitempty"`
	PersonSeniorities     []string `json:"person_seniorities,omitempty"`
	Page                  int      `json:"page,omitempty"`
	PerPage               int      `json:"per_page,omitempty"`
}

type Person struct {
	ID           string          `json:"id"`
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	Name         string          `json:"name"`
	Title        string          `json:"title"`
	LinkedInURL  string          `json:"linkedin_url"`
	Email        string          `json:"email"`
	EmailStatus  string          `json:"email_status"`
	City         string          `json:"city"`
	State        string          `json:"state"`
	Country      string          `json:"country"`
	Organization Organization    `json:"organization"`
	Raw          json.RawMessage `json:"-"`
}

type Organization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	WebsiteURL  string `json:"website_url"`
	LinkedInURL string `json:"linkedin_url"`
	Domain      string `json:"primary_domain"`
}

type Pagination struct {
	Page         int `json:"page"`
	PerPage      int `json:"per_page"`
	TotalEntries int `json:"total_entries"`
	TotalPages   int `json:"total_pages"`
}

type SearchResponse struct {
	People     []Person   `json:"people"`
	Pagination Pagination `json:"pagination"`
	Raw        []byte     `json:"-"`
}

type EnrichPerson struct {
	ID          string `json:"id,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	Domain      string `json:"domain,omitempty"`
	CompanyName string `json:"organization_name,omitempty"`
	LinkedInURL string `json:"linkedin_url,omitempty"`
}

type BulkEnrichRequest struct {
	Details []EnrichPerson `json:"details"`
}

type BulkEnrichResponse struct {
	People []Person `json:"people"`
	Raw    []byte   `json:"-"`
}

type UsageStatsResponse struct {
	Raw []byte `json:"-"`
}
