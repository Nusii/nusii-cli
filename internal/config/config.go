package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	DefaultAPIURL = "https://app.nusii.com"
	DefaultOutput = "table"
)

type Config struct {
	APIKey  string `mapstructure:"api_key"`
	APIURL  string `mapstructure:"api_url"`
	Output  string `mapstructure:"output"`
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

	return v.WriteConfigAs(ConfigPath())
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
