package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestCallbackServerSuccess(t *testing.T) {
	state := "test-state-123"
	resultCh, err := StartCallbackServer(state, 10*time.Second)
	if err != nil {
		t.Fatalf("failed to start callback server: %v", err)
	}

	// Simulate browser callback
	callbackURL := fmt.Sprintf("http://%s/callback?code=auth-code-xyz&state=%s", callbackAddr, state)
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("callback request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	result := <-resultCh
	if result.Error != "" {
		t.Errorf("unexpected error: %s", result.Error)
	}
	if result.Code != "auth-code-xyz" {
		t.Errorf("expected code auth-code-xyz, got %s", result.Code)
	}
}

func TestCallbackServerStateMismatch(t *testing.T) {
	// Wait for previous test's server to shut down
	time.Sleep(100 * time.Millisecond)

	resultCh, err := StartCallbackServer("expected-state", 10*time.Second)
	if err != nil {
		t.Fatalf("failed to start callback server: %v", err)
	}

	callbackURL := fmt.Sprintf("http://%s/callback?code=auth-code&state=wrong-state", callbackAddr)
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("callback request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	result := <-resultCh
	if result.Error != "state mismatch" {
		t.Errorf("expected state mismatch error, got %q", result.Error)
	}
}

func TestCallbackServerErrorParam(t *testing.T) {
	time.Sleep(100 * time.Millisecond)

	resultCh, err := StartCallbackServer("state", 10*time.Second)
	if err != nil {
		t.Fatalf("failed to start callback server: %v", err)
	}

	callbackURL := fmt.Sprintf("http://%s/callback?error=access_denied&error_description=User+denied+access", callbackAddr)
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("callback request failed: %v", err)
	}
	resp.Body.Close()

	result := <-resultCh
	if result.Error != "User denied access" {
		t.Errorf("expected 'User denied access', got %q", result.Error)
	}
}

func TestRedirectURI(t *testing.T) {
	expected := "http://127.0.0.1:18192/callback"
	if got := RedirectURI(); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
