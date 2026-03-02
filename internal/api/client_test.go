package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nusii/nusii-cli/internal/auth"
)

func TestNewClient(t *testing.T) {
	c := NewClient("https://app.nusii.com", nil)
	if c.BaseURL != "https://app.nusii.com/api/v2" {
		t.Errorf("expected base URL to include /api/v2, got %s", c.BaseURL)
	}

	c2 := NewClient("https://app.nusii.com/api/v2", nil)
	if c2.BaseURL != "https://app.nusii.com/api/v2" {
		t.Errorf("expected base URL to remain unchanged, got %s", c2.BaseURL)
	}
}

func TestAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Token token=test-key" {
			t.Errorf("expected auth header 'Token token=test-key', got '%s'", auth)
		}
		ua := r.Header.Get("User-Agent")
		if ua != "nusii-cli" {
			t.Errorf("expected user-agent 'nusii-cli', got '%s'", ua)
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"data":{}}`))
	}))
	defer server.Close()

	a := auth.NewTokenAuth("test-key")
	c := NewClient(server.URL, a)
	resp, err := c.Get("/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
}

func TestErrorParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, auth.NewTokenAuth("key"))
	_, err := c.ReadRawBody(mustGet(t, c, "/test"))
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
	if apiErr.ExitCode() != 3 {
		t.Errorf("expected exit code 3, got %d", apiErr.ExitCode())
	}
}

func TestExitCodes(t *testing.T) {
	tests := []struct {
		status   int
		exitCode int
	}{
		{401, 2},
		{404, 3},
		{422, 4},
		{429, 5},
		{500, 1},
	}
	for _, tt := range tests {
		e := &APIError{StatusCode: tt.status}
		if e.ExitCode() != tt.exitCode {
			t.Errorf("status %d: expected exit code %d, got %d", tt.status, tt.exitCode, e.ExitCode())
		}
	}
}

func TestPaginatedPath(t *testing.T) {
	path := buildPaginatedPath("/clients", 2, 10, nil)
	if path != "/clients?page=2&per_page=10" {
		t.Errorf("unexpected path: %s", path)
	}

	path2 := buildPaginatedPath("/proposals", 1, 0, map[string]string{"status": "draft"})
	// Should contain page=1 and status=draft
	if path2 != "/proposals?page=1&status=draft" {
		t.Errorf("unexpected path: %s", path2)
	}

	path3 := buildPaginatedPath("/clients", 0, 0, nil)
	if path3 != "/clients" {
		t.Errorf("expected no query params, got: %s", path3)
	}
}

func TestListClients(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/clients" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":   "1",
					"type": "clients",
					"attributes": map[string]interface{}{
						"name":  "John",
						"email": "john@example.com",
					},
				},
			},
			"meta": map[string]interface{}{
				"current_page": 1,
				"total_pages":  1,
				"total_count":  1,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient(server.URL, auth.NewTokenAuth("key"))
	_, result, err := c.ListClients(0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 client, got %d", len(result.Data))
	}
	if result.Data[0].Attributes.Name != "John" {
		t.Errorf("expected name 'John', got '%s'", result.Data[0].Attributes.Name)
	}
	if result.Meta.TotalCount != 1 {
		t.Errorf("expected total count 1, got %d", result.Meta.TotalCount)
	}
}

func TestGetAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/account/me" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "1",
				"type": "accounts",
				"attributes": map[string]interface{}{
					"name":  "Test Account",
					"email": "test@example.com",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient(server.URL, auth.NewTokenAuth("key"))
	_, result, err := c.GetAccount()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Data.Attributes.Name != "Test Account" {
		t.Errorf("expected 'Test Account', got '%s'", result.Data.Attributes.Name)
	}
}

func TestRateLimitRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"data":{}}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, auth.NewTokenAuth("key"))
	resp, err := c.Get("/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func mustGet(t *testing.T, c *Client, path string) *http.Response {
	t.Helper()
	resp, err := c.Get(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return resp
}
