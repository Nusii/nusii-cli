package cmd

const (
	oauthCallbackPort      = "18192"
	oauthScopes            = "read write"
	devFallbackClientID    = "-YM-udb7npKuFrhmFO-955qe9IROKr0r4ganH9Zki7Y"
)

// resolveOAuthClientID returns the OAuth client ID with priority:
// config file > env var > build-time ldflags > dev fallback.
func resolveOAuthClientID() string {
	if cfg != nil && cfg.OAuthClientID != "" {
		return cfg.OAuthClientID
	}
	if oauthClientID != "" {
		return oauthClientID
	}
	return devFallbackClientID
}
