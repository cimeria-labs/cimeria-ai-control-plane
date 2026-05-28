package apollo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	cfg  Config
	http HTTPDoer
}

func NewClient(cfg Config, doer HTTPDoer) *Client {
	if doer == nil {
		doer = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{cfg: cfg, http: doer}
}

func NewClientFromEnv() *Client {
	return NewClient(Config{
		APIKey:  strings.TrimSpace(os.Getenv("APOLLO_API_KEY")),
		BaseURL: strings.TrimSpace(os.Getenv("APOLLO_BASE_URL")),
	}, nil)
}

func (c *Client) Configured() bool {
	return c != nil && c.cfg.Configured()
}

func (c *Client) SearchPeople(ctx context.Context, req SearchRequest) (SearchResponse, error) {
	if err := c.requireConfigured(); err != nil {
		return SearchResponse{}, err
	}
	values := url.Values{}
	addList(values, "person_titles[]", req.PersonTitles)
	addList(values, "person_locations[]", req.PersonLocations)
	addList(values, "organization_locations[]", req.OrganizationLocations)
	addList(values, "q_organization_keyword_tags[]", req.OrganizationKeywords)
	addList(values, "person_seniorities[]", req.PersonSeniorities)
	if req.Page > 0 {
		values.Set("page", strconv.Itoa(req.Page))
	}
	if req.PerPage > 0 {
		values.Set("per_page", strconv.Itoa(req.PerPage))
	}

	var out SearchResponse
	raw, err := c.doJSON(ctx, http.MethodPost, "/api/v1/mixed_people/api_search", values, nil)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("decode apollo people search: %w", err)
	}
	out.Raw = raw
	return out, nil
}

func (c *Client) BulkEnrichPeople(ctx context.Context, req BulkEnrichRequest) (BulkEnrichResponse, error) {
	if err := c.requireConfigured(); err != nil {
		return BulkEnrichResponse{}, err
	}
	if len(req.Details) == 0 {
		return BulkEnrichResponse{}, errors.New("apollo enrichment requires at least one person")
	}
	if len(req.Details) > 10 {
		return BulkEnrichResponse{}, errors.New("apollo bulk enrichment accepts at most 10 people")
	}

	values := url.Values{}
	values.Set("reveal_personal_emails", "false")
	values.Set("reveal_phone_number", "false")
	values.Set("run_waterfall_email", "false")
	values.Set("run_waterfall_phone", "false")

	var out BulkEnrichResponse
	raw, err := c.doJSON(ctx, http.MethodPost, "/api/v1/people/bulk_match", values, req)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("decode apollo bulk enrichment: %w", err)
	}
	out.Raw = raw
	return out, nil
}

func (c *Client) UsageStats(ctx context.Context) (UsageStatsResponse, error) {
	if err := c.requireConfigured(); err != nil {
		return UsageStatsResponse{}, err
	}
	raw, err := c.doJSON(ctx, http.MethodPost, "/api/v1/usage_stats/api_usage_stats", nil, nil)
	if err != nil {
		return UsageStatsResponse{}, err
	}
	return UsageStatsResponse{Raw: raw}, nil
}

func (c *Client) requireConfigured() error {
	if c == nil || !c.cfg.Configured() {
		return errors.New("apollo api key is not configured")
	}
	return nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, values url.Values, body any) ([]byte, error) {
	u, err := url.Parse(strings.TrimRight(c.cfg.EffectiveBaseURL(), "/") + path)
	if err != nil {
		return nil, err
	}
	if values != nil {
		u.RawQuery = values.Encode()
	}

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("X-Api-Key", c.cfg.APIKey)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("apollo api returned %d", res.StatusCode)
	}
	return raw, nil
}

func addList(values url.Values, key string, items []string) {
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			values.Add(key, item)
		}
	}
}
