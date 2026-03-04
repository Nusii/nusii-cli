package config

import (
	"os"
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	// Isolate from real config file
	os.Setenv("XDG_CONFIG_HOME", t.TempDir())
	defer os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("NUSII_API_KEY")
	os.Unsetenv("NUSII_API_URL")
	os.Unsetenv("NUSII_OUTPUT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.APIURL != DefaultAPIURL {
		t.Errorf("expected default API URL %s, got %s", DefaultAPIURL, cfg.APIURL)
	}
	if cfg.Output != DefaultOutput {
		t.Errorf("expected default output %s, got %s", DefaultOutput, cfg.Output)
	}
}

func TestEnvOverrides(t *testing.T) {
	os.Setenv("NUSII_API_KEY", "env-key")
	os.Setenv("NUSII_API_URL", "http://localhost:3000")
	os.Setenv("NUSII_OUTPUT", "json")
	defer func() {
		os.Unsetenv("NUSII_API_KEY")
		os.Unsetenv("NUSII_API_URL")
		os.Unsetenv("NUSII_OUTPUT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.APIKey != "env-key" {
		t.Errorf("expected API key 'env-key', got '%s'", cfg.APIKey)
	}
	if cfg.APIURL != "http://localhost:3000" {
		t.Errorf("expected API URL 'http://localhost:3000', got '%s'", cfg.APIURL)
	}
	if cfg.Output != "json" {
		t.Errorf("expected output 'json', got '%s'", cfg.Output)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Use temp dir for config
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("NUSII_API_KEY")
	os.Unsetenv("NUSII_API_URL")
	os.Unsetenv("NUSII_OUTPUT")

	cfg := &Config{
		APIKey: "saved-key",
		APIURL: "https://custom.nusii.com",
		Output: "json",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if loaded.APIKey != "saved-key" {
		t.Errorf("expected API key 'saved-key', got '%s'", loaded.APIKey)
	}
	if loaded.APIURL != "https://custom.nusii.com" {
		t.Errorf("expected API URL 'https://custom.nusii.com', got '%s'", loaded.APIURL)
	}
}

func TestRemoveAPIKey(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("NUSII_API_KEY")

	cfg := &Config{
		APIKey: "to-remove",
		APIURL: DefaultAPIURL,
		Output: DefaultOutput,
	}
	Save(cfg)

	if err := RemoveAPIKey(); err != nil {
		t.Fatalf("remove error: %v", err)
	}

	loaded, _ := Load()
	if loaded.APIKey != "" {
		t.Errorf("expected empty API key after removal, got '%s'", loaded.APIKey)
	}
}

func TestHasOAuthToken(t *testing.T) {
	cfg := &Config{}
	if cfg.HasOAuthToken() {
		t.Error("expected HasOAuthToken false with empty config")
	}

	cfg.OAuthAccessToken = "access"
	if cfg.HasOAuthToken() {
		t.Error("expected HasOAuthToken false with only access token")
	}

	cfg.OAuthRefreshToken = "refresh"
	if !cfg.HasOAuthToken() {
		t.Error("expected HasOAuthToken true with both tokens")
	}
}

func TestAuthMethod(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		expected string
	}{
		{"none", Config{}, "none"},
		{"api_key", Config{APIKey: "key"}, "api_key"},
		{"oauth", Config{OAuthAccessToken: "a", OAuthRefreshToken: "r"}, "oauth"},
		{"oauth_over_api_key", Config{APIKey: "key", OAuthAccessToken: "a", OAuthRefreshToken: "r"}, "oauth"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.AuthMethod(); got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestOAuthExpiryTime(t *testing.T) {
	cfg := &Config{}
	ts, err := cfg.OAuthExpiryTime()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ts.IsZero() {
		t.Error("expected zero time for empty expiry")
	}

	now := time.Now().Truncate(time.Second)
	cfg.OAuthExpiry = now.Format(time.RFC3339)
	ts, err = cfg.OAuthExpiryTime()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ts.Equal(now) {
		t.Errorf("expected %v, got %v", now, ts)
	}
}

func TestSaveOAuthTokens(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("NUSII_API_KEY")

	// Start with an API key
	initial := &Config{
		APIKey: "old-key",
		APIURL: DefaultAPIURL,
		Output: DefaultOutput,
	}
	Save(initial)

	expiry := time.Now().Add(2 * time.Hour).Truncate(time.Second)
	if err := SaveOAuthTokens("access-tok", "refresh-tok", expiry, "client-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := Load()
	if loaded.APIKey != "" {
		t.Errorf("expected API key cleared, got '%s'", loaded.APIKey)
	}
	if loaded.OAuthAccessToken != "access-tok" {
		t.Errorf("expected access-tok, got '%s'", loaded.OAuthAccessToken)
	}
	if loaded.OAuthRefreshToken != "refresh-tok" {
		t.Errorf("expected refresh-tok, got '%s'", loaded.OAuthRefreshToken)
	}
	if loaded.OAuthClientID != "client-123" {
		t.Errorf("expected client-123, got '%s'", loaded.OAuthClientID)
	}
}

func TestClearAuth(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("NUSII_API_KEY")

	initial := &Config{
		APIKey:            "key",
		APIURL:            DefaultAPIURL,
		Output:            DefaultOutput,
		OAuthAccessToken:  "access",
		OAuthRefreshToken: "refresh",
		OAuthExpiry:       "2025-01-01T00:00:00Z",
		OAuthClientID:     "client",
	}
	Save(initial)

	if err := ClearAuth(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := Load()
	if loaded.APIKey != "" {
		t.Errorf("expected empty API key, got '%s'", loaded.APIKey)
	}
	if loaded.OAuthAccessToken != "" {
		t.Errorf("expected empty access token, got '%s'", loaded.OAuthAccessToken)
	}
	if loaded.OAuthRefreshToken != "" {
		t.Errorf("expected empty refresh token, got '%s'", loaded.OAuthRefreshToken)
	}
	if loaded.OAuthClientID != "" {
		t.Errorf("expected empty client ID, got '%s'", loaded.OAuthClientID)
	}
}
