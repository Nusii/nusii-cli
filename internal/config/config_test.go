package config

import (
	"os"
	"testing"
)

func TestDefaults(t *testing.T) {
	// Clear env vars that might interfere
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
