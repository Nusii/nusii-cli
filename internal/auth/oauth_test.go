package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOAuthAuth_Apply(t *testing.T) {
	o := &OAuthAuth{
		AccessToken: "test-access-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	if err := o.Apply(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := req.Header.Get("Authorization")
	expected := "Bearer test-access-token"
	if got != expected {
		t.Errorf("expected header %q, got %q", expected, got)
	}
}

func TestOAuthAuth_ApplyEmpty(t *testing.T) {
	o := &OAuthAuth{}
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	err := o.Apply(req)
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestOAuthAuth_AutoRefresh(t *testing.T) {
	refreshCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshCalled = true
		if r.FormValue("grant_type") != "refresh_token" {
			t.Errorf("expected grant_type=refresh_token, got %s", r.FormValue("grant_type"))
		}
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    7200,
		})
	}))
	defer server.Close()

	savedAccess := ""
	savedRefresh := ""
	o := &OAuthAuth{
		AccessToken:  "old-access-token",
		RefreshToken: "old-refresh-token",
		Expiry:       time.Now().Add(-1 * time.Minute), // expired
		TokenURL:     server.URL,
		ClientID:     "test-client",
		OnRefresh: func(access, refresh string, expiry time.Time) error {
			savedAccess = access
			savedRefresh = refresh
			return nil
		},
	}

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	if err := o.Apply(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !refreshCalled {
		t.Error("expected refresh to be called")
	}

	if o.AccessToken != "new-access-token" {
		t.Errorf("expected new-access-token, got %s", o.AccessToken)
	}
	if o.RefreshToken != "new-refresh-token" {
		t.Errorf("expected new-refresh-token, got %s", o.RefreshToken)
	}
	if savedAccess != "new-access-token" {
		t.Errorf("OnRefresh not called with new access token")
	}
	if savedRefresh != "new-refresh-token" {
		t.Errorf("OnRefresh not called with new refresh token")
	}

	got := req.Header.Get("Authorization")
	if got != "Bearer new-access-token" {
		t.Errorf("expected Bearer new-access-token, got %s", got)
	}
}

func TestExchangeCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("grant_type") != "authorization_code" {
			t.Errorf("expected grant_type=authorization_code, got %s", r.FormValue("grant_type"))
		}
		if r.FormValue("code") != "test-code" {
			t.Errorf("expected code=test-code, got %s", r.FormValue("code"))
		}
		if r.FormValue("code_verifier") != "test-verifier" {
			t.Errorf("expected code_verifier=test-verifier, got %s", r.FormValue("code_verifier"))
		}
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken:  "exchanged-token",
			RefreshToken: "exchanged-refresh",
			ExpiresIn:    7200,
		})
	}))
	defer server.Close()

	resp, err := ExchangeCode(server.URL, "test-client", "test-code", "test-verifier", "http://127.0.0.1:18192/callback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.AccessToken != "exchanged-token" {
		t.Errorf("expected exchanged-token, got %s", resp.AccessToken)
	}
	if resp.RefreshToken != "exchanged-refresh" {
		t.Errorf("expected exchanged-refresh, got %s", resp.RefreshToken)
	}
}

func TestExchangeCodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()

	_, err := ExchangeCode(server.URL, "test-client", "bad-code", "verifier", "http://127.0.0.1:18192/callback")
	if err == nil {
		t.Fatal("expected error for bad code exchange")
	}
}

func TestRevokeToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("token") != "revoke-me" {
			t.Errorf("expected token=revoke-me, got %s", r.FormValue("token"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := RevokeToken(server.URL, "test-client", "revoke-me")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
