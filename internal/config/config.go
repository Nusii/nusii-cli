package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

const (
	DefaultAPIURL = "https://app.nusii.com"
	DefaultOutput = "table"
)

type Config struct {
	APIKey            string `mapstructure:"api_key"`
	APIURL            string `mapstructure:"api_url"`
	Output            string `mapstructure:"output"`
	OAuthAccessToken  string `mapstructure:"oauth_access_token"`
	OAuthRefreshToken string `mapstructure:"oauth_refresh_token"`
	OAuthExpiry       string `mapstructure:"oauth_expiry"`
	OAuthClientID     string `mapstructure:"oauth_client_id"`
}

// ConfigDir returns the nusii config directory path.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "nusii")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "nusii")
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// Load reads the config from file, env vars, and applies defaults.
func Load() (*Config, error) {
	v := viper.New()

	v.SetDefault("api_url", DefaultAPIURL)
	v.SetDefault("output", DefaultOutput)

	// Config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(ConfigDir())
	_ = v.ReadInConfig() // ok if missing

	// Environment variables
	v.SetEnvPrefix("NUSII")
	v.AutomaticEnv()
	v.BindEnv("api_key", "NUSII_API_KEY")
	v.BindEnv("api_url", "NUSII_API_URL")
	v.BindEnv("output", "NUSII_OUTPUT")
	v.BindEnv("oauth_client_id", "NUSII_OAUTH_CLIENT_ID")

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes the config to the config file.
func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.Set("api_key", cfg.APIKey)
	v.Set("api_url", cfg.APIURL)
	v.Set("output", cfg.Output)
	v.Set("oauth_access_token", cfg.OAuthAccessToken)
	v.Set("oauth_refresh_token", cfg.OAuthRefreshToken)
	v.Set("oauth_expiry", cfg.OAuthExpiry)
	v.Set("oauth_client_id", cfg.OAuthClientID)

	return v.WriteConfigAs(ConfigPath())
}

// HasOAuthToken returns true if OAuth tokens are configured.
func (c *Config) HasOAuthToken() bool {
	return c.OAuthAccessToken != "" && c.OAuthRefreshToken != ""
}

// AuthMethod returns the current authentication method: "oauth", "api_key", or "none".
func (c *Config) AuthMethod() string {
	if c.HasOAuthToken() {
		return "oauth"
	}
	if c.APIKey != "" {
		return "api_key"
	}
	return "none"
}

// OAuthExpiryTime parses the stored OAuth expiry as a time.Time.
func (c *Config) OAuthExpiryTime() (time.Time, error) {
	if c.OAuthExpiry == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, c.OAuthExpiry)
}

// SaveOAuthTokens stores OAuth tokens and clears the API key.
func SaveOAuthTokens(accessToken, refreshToken string, expiry time.Time, clientID string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.APIKey = ""
	cfg.OAuthAccessToken = accessToken
	cfg.OAuthRefreshToken = refreshToken
	cfg.OAuthExpiry = expiry.Format(time.RFC3339)
	cfg.OAuthClientID = clientID
	return Save(cfg)
}

// ClearAuth removes both API key and OAuth tokens from the config.
func ClearAuth() error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.APIKey = ""
	cfg.OAuthAccessToken = ""
	cfg.OAuthRefreshToken = ""
	cfg.OAuthExpiry = ""
	cfg.OAuthClientID = ""
	return Save(cfg)
}

// RemoveAPIKey removes the API key from the config file.
func RemoveAPIKey() error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.APIKey = ""
	return Save(cfg)
}
