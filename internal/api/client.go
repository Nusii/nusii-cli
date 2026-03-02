package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nusii/nusii-cli/internal/auth"
	"github.com/nusii/nusii-cli/internal/models"
)

const userAgent = "nusii-cli"

// Client is the Nusii API HTTP client.
type Client struct {
	BaseURL string
	Auth    auth.Authenticator
	HTTP    *http.Client
	Debug   bool
	Version string
}

// NewClient creates a new API client.
func NewClient(baseURL string, authenticator auth.Authenticator) *Client {
	if !strings.HasSuffix(baseURL, "/api/v2") {
		baseURL = strings.TrimRight(baseURL, "/") + "/api/v2"
	}
	return &Client{
		BaseURL: baseURL,
		Auth:    authenticator,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// APIError represents a structured API error.
type APIError struct {
	StatusCode int    `json:"status"`
	Message    string `json:"error"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// ExitCode returns the appropriate CLI exit code for the error.
func (e *APIError) ExitCode() int {
	switch e.StatusCode {
	case 401:
		return 2
	case 404:
		return 3
	case 422:
		return 4
	case 429:
		return 5
	default:
		return 1
	}
}

func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	u := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, u, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	ua := userAgent
	if c.Version != "" {
		ua += "/" + c.Version
	}
	req.Header.Set("User-Agent", ua)

	if c.Auth != nil {
		if err := c.Auth.Apply(req); err != nil {
			return nil, err
		}
	}

	if c.Debug {
		debugRequest(req)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}

	if c.Debug {
		debugResponse(resp)
	}

	// Handle rate limiting
	if resp.StatusCode == 429 {
		retryAfter := resp.Header.Get("Retry-After")
		seconds := 5
		if retryAfter != "" {
			if s, err := strconv.Atoi(retryAfter); err == nil {
				seconds = s
			}
		}
		resp.Body.Close()
		time.Sleep(time.Duration(seconds) * time.Second)
		return c.doRequest(method, path, body)
	}

	return resp, nil
}

func (c *Client) parseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(bodyBytes))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return &APIError{StatusCode: resp.StatusCode, Message: msg}
	}

	if target == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// ReadRawBody reads the response body as raw bytes (for JSON passthrough).
func (c *Client) ReadRawBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(bodyBytes))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Message: msg}
	}
	return io.ReadAll(resp.Body)
}

// Get performs a GET request.
func (c *Client) Get(path string) (*http.Response, error) {
	return c.doRequest("GET", path, nil)
}

// Post performs a POST request.
func (c *Client) Post(path string, body interface{}) (*http.Response, error) {
	return c.doRequest("POST", path, body)
}

// Put performs a PUT request.
func (c *Client) Put(path string, body interface{}) (*http.Response, error) {
	return c.doRequest("PUT", path, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) (*http.Response, error) {
	return c.doRequest("DELETE", path, nil)
}

// Pagination helpers

func buildPaginatedPath(base string, page, perPage int, extra map[string]string) string {
	params := url.Values{}
	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		params.Set("per_page", strconv.Itoa(perPage))
	}
	for k, v := range extra {
		if v != "" {
			params.Set(k, v)
		}
	}
	if len(params) > 0 {
		return base + "?" + params.Encode()
	}
	return base
}

// Resource-specific methods

// GetAccount fetches the current account.
func (c *Client) GetAccount() ([]byte, *models.SingleResponse[models.Account], error) {
	resp, err := c.Get("/account/me")
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.Account]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// ListClients fetches a paginated list of clients.
func (c *Client) ListClients(page, perPage int) ([]byte, *models.ListResponse[models.Client], error) {
	path := buildPaginatedPath("/clients", page, perPage, nil)
	resp, err := c.Get(path)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.ListResponse[models.Client]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// GetClient fetches a single client.
func (c *Client) GetClient(id string) ([]byte, *models.SingleResponse[models.Client], error) {
	resp, err := c.Get("/clients/" + id)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.Client]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// CreateClient creates a new client.
func (c *Client) CreateClient(attrs models.Client) ([]byte, *models.SingleResponse[models.Client], error) {
	body := map[string]models.Client{"client": attrs}
	resp, err := c.Post("/clients", body)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.Client]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// UpdateClient updates an existing client.
func (c *Client) UpdateClient(id string, attrs models.Client) ([]byte, *models.SingleResponse[models.Client], error) {
	body := map[string]models.Client{"client": attrs}
	resp, err := c.Put("/clients/"+id, body)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.Client]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// DeleteClient deletes a client.
func (c *Client) DeleteClient(id string) error {
	resp, err := c.Delete("/clients/" + id)
	if err != nil {
		return err
	}
	return c.parseResponse(resp, nil)
}

// ListProposals fetches a paginated list of proposals.
func (c *Client) ListProposals(page, perPage int, status string, archived bool) ([]byte, *models.ListResponse[models.Proposal], error) {
	extra := map[string]string{}
	if status != "" {
		extra["status"] = status
	}
	if archived {
		extra["archived"] = "true"
	}
	path := buildPaginatedPath("/proposals", page, perPage, extra)
	resp, err := c.Get(path)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.ListResponse[models.Proposal]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// GetProposal fetches a single proposal.
func (c *Client) GetProposal(id string) ([]byte, *models.SingleResponse[models.Proposal], error) {
	resp, err := c.Get("/proposals/" + id)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.Proposal]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// CreateProposal creates a new proposal.
func (c *Client) CreateProposal(attrs models.Proposal) ([]byte, *models.SingleResponse[models.Proposal], error) {
	body := map[string]models.Proposal{"proposal": attrs}
	resp, err := c.Post("/proposals", body)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.Proposal]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// UpdateProposal updates an existing proposal.
func (c *Client) UpdateProposal(id string, attrs models.Proposal) ([]byte, *models.SingleResponse[models.Proposal], error) {
	body := map[string]models.Proposal{"proposal": attrs}
	resp, err := c.Put("/proposals/"+id, body)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.Proposal]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// DeleteProposal deletes a proposal.
func (c *Client) DeleteProposal(id string) error {
	resp, err := c.Delete("/proposals/" + id)
	if err != nil {
		return err
	}
	return c.parseResponse(resp, nil)
}

// SendProposal sends a proposal.
func (c *Client) SendProposal(id string, sendReq models.ProposalSendRequest) ([]byte, error) {
	resp, err := c.Put("/proposals/"+id+"/send_proposal", sendReq)
	if err != nil {
		return nil, err
	}
	return c.ReadRawBody(resp)
}

// ArchiveProposal archives a proposal.
func (c *Client) ArchiveProposal(id string) ([]byte, error) {
	resp, err := c.Put("/proposals/"+id+"/archive", nil)
	if err != nil {
		return nil, err
	}
	return c.ReadRawBody(resp)
}

// ListSections fetches sections, optionally filtered.
func (c *Client) ListSections(page, perPage int, proposalID, templateID string, includeLineItems bool) ([]byte, *models.ListResponse[models.Section], error) {
	extra := map[string]string{}
	if proposalID != "" {
		extra["proposal_id"] = proposalID
	}
	if templateID != "" {
		extra["template_id"] = templateID
	}
	if includeLineItems {
		extra["include_line_items"] = "true"
	}
	path := buildPaginatedPath("/sections", page, perPage, extra)
	resp, err := c.Get(path)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.ListResponse[models.Section]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// GetSection fetches a single section.
func (c *Client) GetSection(id string) ([]byte, *models.SingleResponse[models.Section], error) {
	resp, err := c.Get("/sections/" + id)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.Section]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// CreateSection creates a new section.
func (c *Client) CreateSection(attrs models.Section) ([]byte, *models.SingleResponse[models.Section], error) {
	body := map[string]models.Section{"section": attrs}
	resp, err := c.Post("/sections", body)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.Section]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// UpdateSection updates an existing section.
func (c *Client) UpdateSection(id string, attrs models.Section) ([]byte, *models.SingleResponse[models.Section], error) {
	body := map[string]models.Section{"section": attrs}
	resp, err := c.Put("/sections/"+id, body)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.Section]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// DeleteSection deletes a section.
func (c *Client) DeleteSection(id string) error {
	resp, err := c.Delete("/sections/" + id)
	if err != nil {
		return err
	}
	return c.parseResponse(resp, nil)
}

// ListLineItems fetches line items, optionally by section.
func (c *Client) ListLineItems(page, perPage int, sectionID string) ([]byte, *models.ListResponse[models.LineItem], error) {
	var basePath string
	if sectionID != "" {
		basePath = "/sections/" + sectionID + "/line_items"
	} else {
		basePath = "/line_items"
	}
	path := buildPaginatedPath(basePath, page, perPage, nil)
	resp, err := c.Get(path)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.ListResponse[models.LineItem]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// GetLineItem fetches a single line item.
func (c *Client) GetLineItem(id string) ([]byte, *models.SingleResponse[models.LineItem], error) {
	resp, err := c.Get("/line_items/" + id)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.LineItem]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// CreateLineItem creates a new line item under a section.
func (c *Client) CreateLineItem(sectionID string, attrs models.LineItem) ([]byte, *models.SingleResponse[models.LineItem], error) {
	body := map[string]models.LineItem{"line_item": attrs}
	resp, err := c.Post("/sections/"+sectionID+"/line_items", body)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.LineItem]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// UpdateLineItem updates an existing line item.
func (c *Client) UpdateLineItem(id string, attrs models.LineItem) ([]byte, *models.SingleResponse[models.LineItem], error) {
	body := map[string]models.LineItem{"line_item": attrs}
	resp, err := c.Put("/line_items/"+id, body)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.LineItem]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// DeleteLineItem deletes a line item.
func (c *Client) DeleteLineItem(id string) error {
	resp, err := c.Delete("/line_items/" + id)
	if err != nil {
		return err
	}
	return c.parseResponse(resp, nil)
}

// ListActivities fetches proposal activities.
func (c *Client) ListActivities(page, perPage int, proposalID, clientID string) ([]byte, *models.ListResponse[models.Activity], error) {
	extra := map[string]string{}
	if proposalID != "" {
		extra["proposal_id"] = proposalID
	}
	if clientID != "" {
		extra["client_id"] = clientID
	}
	path := buildPaginatedPath("/proposal_activities", page, perPage, extra)
	resp, err := c.Get(path)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.ListResponse[models.Activity]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// GetActivity fetches a single activity.
func (c *Client) GetActivity(id string) ([]byte, *models.SingleResponse[models.Activity], error) {
	resp, err := c.Get("/proposal_activities/" + id)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.Activity]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// ListUsers fetches users.
func (c *Client) ListUsers(page, perPage int) ([]byte, *models.ListResponse[models.User], error) {
	path := buildPaginatedPath("/users", page, perPage, nil)
	resp, err := c.Get(path)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.ListResponse[models.User]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// ThemeItem represents a theme as returned by the API (plain object, not JSON:API).
type ThemeItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListThemes fetches themes. The API returns a plain array, not JSON:API format.
func (c *Client) ListThemes() ([]byte, []ThemeItem, error) {
	resp, err := c.Get("/themes")
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result []ThemeItem
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, result, nil
}

// ListWebhooks fetches webhook endpoints.
func (c *Client) ListWebhooks(page, perPage int) ([]byte, *models.ListResponse[models.WebhookEndpoint], error) {
	path := buildPaginatedPath("/webhook_endpoints", page, perPage, nil)
	resp, err := c.Get(path)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.ListResponse[models.WebhookEndpoint]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// GetWebhook fetches a single webhook endpoint.
func (c *Client) GetWebhook(id string) ([]byte, *models.SingleResponse[models.WebhookEndpoint], error) {
	resp, err := c.Get("/webhook_endpoints/" + id)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.WebhookEndpoint]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// CreateWebhook creates a new webhook endpoint.
func (c *Client) CreateWebhook(attrs models.WebhookEndpoint) ([]byte, *models.SingleResponse[models.WebhookEndpoint], error) {
	body := map[string]models.WebhookEndpoint{"webhook_endpoint": attrs}
	resp, err := c.Post("/webhook_endpoints", body)
	if err != nil {
		return nil, nil, err
	}
	raw, err := c.ReadRawBody(resp)
	if err != nil {
		return nil, nil, err
	}
	var result models.SingleResponse[models.WebhookEndpoint]
	if err := json.Unmarshal(raw, &result); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

// DeleteWebhook deletes a webhook endpoint.
func (c *Client) DeleteWebhook(id string) error {
	resp, err := c.Delete("/webhook_endpoints/" + id)
	if err != nil {
		return err
	}
	return c.parseResponse(resp, nil)
}

// debug helpers

func debugRequest(req *http.Request) {
	fmt.Fprintf(os.Stderr, "→ %s %s\n", req.Method, req.URL)
	for k, v := range req.Header {
		if k == "Authorization" {
			fmt.Fprintf(os.Stderr, "  %s: [REDACTED]\n", k)
		} else {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", k, strings.Join(v, ", "))
		}
	}
}

func debugResponse(resp *http.Response) {
	fmt.Fprintf(os.Stderr, "← %s\n", resp.Status)
	for k, v := range resp.Header {
		fmt.Fprintf(os.Stderr, "  %s: %s\n", k, strings.Join(v, ", "))
	}
}
