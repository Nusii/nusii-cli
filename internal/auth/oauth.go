package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TokenResponse represents the JSON response from the OAuth token endpoint.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreatedAt    int64  `json:"created_at"`
}

// TokenSaver is a callback invoked when tokens are refreshed.
type TokenSaver func(accessToken, refreshToken string, expiry time.Time) error

// OAuthAuth implements Authenticator using OAuth2 Bearer tokens with auto-refresh.
type OAuthAuth struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
	TokenURL     string
	ClientID     string
	OnRefresh    TokenSaver

	mu sync.Mutex
}

func (o *OAuthAuth) Apply(req *http.Request) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.AccessToken == "" {
		return fmt.Errorf("not authenticated. Run 'nusii auth login' to authenticate")
	}

	// Auto-refresh if token expires within 60 seconds
	if !o.Expiry.IsZero() && time.Until(o.Expiry) < 60*time.Second {
		if err := o.refresh(); err != nil {
			return fmt.Errorf("refreshing OAuth token: %w", err)
		}
	}

	req.Header.Set("Authorization", "Bearer "+o.AccessToken)
	return nil
}

func (o *OAuthAuth) refresh() error {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {o.RefreshToken},
		"client_id":     {o.ClientID},
	}

	resp, err := http.PostForm(o.TokenURL, data)
	if err != nil {
		return fmt.Errorf("token refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token refresh failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decoding token response: %w", err)
	}

	o.AccessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		o.RefreshToken = tokenResp.RefreshToken
	}
	o.Expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	if o.OnRefresh != nil {
		if err := o.OnRefresh(o.AccessToken, o.RefreshToken, o.Expiry); err != nil {
			return fmt.Errorf("saving refreshed tokens: %w", err)
		}
	}

	return nil
}

// ExchangeCode exchanges an authorization code + PKCE verifier for tokens.
func ExchangeCode(tokenURL, clientID, code, verifier, redirectURI string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"code_verifier": {verifier},
		"redirect_uri":  {redirectURI},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("token exchange request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}

	return &tokenResp, nil
}

// RevokeToken revokes an OAuth token server-side.
func RevokeToken(revokeURL, clientID, token string) error {
	data := url.Values{
		"client_id": {clientID},
		"token":     {token},
	}

	resp, err := http.PostForm(revokeURL, data)
	if err != nil {
		return fmt.Errorf("token revoke request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token revoke failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}
