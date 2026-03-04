package cmd

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/nusii/nusii-cli/internal/auth"
	"github.com/nusii/nusii-cli/internal/config"
	"github.com/nusii/nusii-cli/internal/output"
)

var (
	flagWithToken bool
	flagWithOAuth bool
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  "Login, check status, or logout from the Nusii API.",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Nusii API",
	Long: `Authenticate with the Nusii API using an API key or OAuth (browser).

Examples:
  nusii auth login
  nusii auth login --with-token
  nusii auth login --with-oauth
  nusii auth login --api-key YOUR_KEY
  nusii auth login --no-input --with-oauth --api-url http://localhost:3000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If --api-key flag was provided, go straight to token flow
		if flagAPIKey != "" {
			return loginWithToken()
		}

		// If explicit method flag, use that
		if flagWithToken {
			return loginWithToken()
		}
		if flagWithOAuth {
			return loginWithOAuth()
		}

		// Non-interactive without explicit method is an error
		if flagNoInput {
			return fmt.Errorf("--no-input requires --with-token or --with-oauth")
		}

		// Interactive menu
		fmt.Fprintln(os.Stderr, "How would you like to authenticate?")
		fmt.Fprintln(os.Stderr, "  1. API Token")
		fmt.Fprintln(os.Stderr, "  2. OAuth (browser)")
		fmt.Fprint(os.Stderr, "Choose [1-2]: ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			return loginWithToken()
		case "2":
			return loginWithOAuth()
		default:
			return fmt.Errorf("invalid choice: %s", choice)
		}
	},
}

func loginWithToken() error {
	apiKey := cfg.APIKey

	// Prompt for API key if not provided
	if apiKey == "" && !flagNoInput {
		fmt.Fprint(os.Stderr, "Enter your Nusii API key: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading API key: %w", err)
		}
		apiKey = strings.TrimSpace(input)
	}

	if apiKey == "" {
		return fmt.Errorf("API key is required. Use --api-key or run interactively")
	}

	// Temporarily set the key to validate
	cfg.APIKey = apiKey
	client := newAPIClient()

	// Validate by fetching account
	_, account, err := client.GetAccount()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Save to config (clears any OAuth tokens)
	saveCfg := &config.Config{
		APIKey: apiKey,
		APIURL: cfg.APIURL,
		Output: cfg.Output,
	}
	if err := config.Save(saveCfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	format := getOutputFormat()
	if format == output.FormatJSON {
		fmt.Printf(`{"status":"authenticated","method":"api_key","account":"%s","email":"%s"}`, account.Data.Attributes.Name, account.Data.Attributes.Email)
		fmt.Println()
	} else {
		fmt.Fprintf(os.Stderr, "Authenticated as %s (%s)\n", account.Data.Attributes.Name, account.Data.Attributes.Email)
		fmt.Fprintf(os.Stderr, "Config saved to %s\n", config.ConfigPath())
	}
	return nil
}

func loginWithOAuth() error {
	clientID := resolveOAuthClientID()
	baseURL := strings.TrimRight(cfg.APIURL, "/")
	// Remove /api/v2 suffix if present for OAuth endpoints
	baseURL = strings.TrimSuffix(baseURL, "/api/v2")
	tokenURL := baseURL + "/oauth/token"
	authorizeURL := baseURL + "/oauth/authorize"
	redirectURI := auth.RedirectURI()

	// Generate PKCE verifier and challenge
	verifier, err := auth.GenerateCodeVerifier()
	if err != nil {
		return fmt.Errorf("generating PKCE verifier: %w", err)
	}
	challenge := auth.GenerateCodeChallenge(verifier)

	// Generate random state
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return fmt.Errorf("generating state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	// Build authorization URL
	params := url.Values{
		"client_id":             {clientID},
		"redirect_uri":         {redirectURI},
		"response_type":        {"code"},
		"scope":                {oauthScopes},
		"state":                {state},
		"code_challenge":       {challenge},
		"code_challenge_method": {"S256"},
	}
	authURL := authorizeURL + "?" + params.Encode()

	// Start callback server
	resultCh, err := auth.StartCallbackServer(state, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("starting callback server: %w", err)
	}

	// Open browser or print URL
	if flagNoInput {
		fmt.Fprintf(os.Stderr, "Open this URL in your browser to authenticate:\n%s\n", authURL)
	} else {
		fmt.Fprintf(os.Stderr, "Opening browser for authentication...\n")
		if err := auth.OpenBrowser(authURL); err != nil {
			fmt.Fprintf(os.Stderr, "Could not open browser. Open this URL manually:\n%s\n", authURL)
		}
	}

	fmt.Fprintf(os.Stderr, "Waiting for authorization...\n")

	// Wait for callback
	result := <-resultCh
	if result.Error != "" {
		return fmt.Errorf("authorization failed: %s", result.Error)
	}

	// Exchange code for tokens
	tokenResp, err := auth.ExchangeCode(tokenURL, clientID, result.Code, verifier, redirectURI)
	if err != nil {
		return fmt.Errorf("exchanging authorization code: %w", err)
	}

	expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Validate by fetching account with the new token
	cfg.OAuthAccessToken = tokenResp.AccessToken
	cfg.OAuthRefreshToken = tokenResp.RefreshToken
	cfg.APIKey = ""
	client := newAPIClient()

	_, account, err := client.GetAccount()
	if err != nil {
		return fmt.Errorf("validating OAuth token: %w", err)
	}

	// Save to config (clears API key, preserves API URL)
	saveCfg := &config.Config{
		APIURL:            cfg.APIURL,
		Output:            cfg.Output,
		OAuthAccessToken:  tokenResp.AccessToken,
		OAuthRefreshToken: tokenResp.RefreshToken,
		OAuthExpiry:       expiry.Format(time.RFC3339),
		OAuthClientID:     clientID,
	}
	if err := config.Save(saveCfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	format := getOutputFormat()
	if format == output.FormatJSON {
		fmt.Printf(`{"status":"authenticated","method":"oauth","account":"%s","email":"%s"}`, account.Data.Attributes.Name, account.Data.Attributes.Email)
		fmt.Println()
	} else {
		fmt.Fprintf(os.Stderr, "Authenticated as %s (%s)\n", account.Data.Attributes.Name, account.Data.Attributes.Email)
		fmt.Fprintf(os.Stderr, "Config saved to %s\n", config.ConfigPath())
	}
	return nil
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	Long: `Show current authentication status and account info.

Examples:
  nusii auth status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		method := cfg.AuthMethod()
		if method == "none" {
			format := getOutputFormat()
			if format == output.FormatJSON {
				fmt.Println(`{"status":"not authenticated"}`)
			} else {
				fmt.Println("Not authenticated. Run 'nusii auth login' to authenticate.")
			}
			return nil
		}

		client := newAPIClient()
		_, account, err := client.GetAccount()
		if err != nil {
			return fmt.Errorf("checking auth status: %w", err)
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			jsonOut := fmt.Sprintf(`{"status":"authenticated","method":"%s","account":"%s","email":"%s","api_url":"%s"`,
				method, account.Data.Attributes.Name, account.Data.Attributes.Email, cfg.APIURL)
			if method == "oauth" {
				if expiry, err := cfg.OAuthExpiryTime(); err == nil && !expiry.IsZero() {
					jsonOut += fmt.Sprintf(`,"token_expiry":"%s"`, expiry.Format(time.RFC3339))
				}
			}
			jsonOut += "}"
			fmt.Println(jsonOut)
		} else {
			fmt.Printf("Authenticated as %s (%s)\n", account.Data.Attributes.Name, account.Data.Attributes.Email)
			fmt.Printf("API URL: %s\n", cfg.APIURL)
			fmt.Printf("Method: %s\n", method)
			if method == "oauth" {
				if expiry, err := cfg.OAuthExpiryTime(); err == nil && !expiry.IsZero() {
					remaining := time.Until(expiry).Truncate(time.Second)
					if remaining > 0 {
						fmt.Printf("Token expires in: %s\n", remaining)
					} else {
						fmt.Printf("Token expired (will refresh on next request)\n")
					}
				}
			}
		}
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored authentication",
	Long: `Remove stored authentication credentials.

For OAuth, this also revokes the token server-side.

Examples:
  nusii auth logout`,
	RunE: func(cmd *cobra.Command, args []string) error {
		method := cfg.AuthMethod()

		// If OAuth, try to revoke the token server-side
		if method == "oauth" && cfg.OAuthAccessToken != "" {
			clientID := resolveOAuthClientID()
			baseURL := strings.TrimSuffix(strings.TrimRight(cfg.APIURL, "/"), "/api/v2")
			revokeURL := baseURL + "/oauth/revoke"
			if err := auth.RevokeToken(revokeURL, clientID, cfg.OAuthAccessToken); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not revoke token server-side: %v\n", err)
			}
		}

		if err := config.ClearAuth(); err != nil {
			return fmt.Errorf("clearing auth: %w", err)
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			fmt.Println(`{"status":"logged out"}`)
		} else {
			if method == "oauth" {
				fmt.Println("Logged out. OAuth tokens revoked and removed from config.")
			} else {
				fmt.Println("Logged out. API key removed from config.")
			}
		}
		return nil
	},
}

func init() {
	authLoginCmd.Flags().BoolVar(&flagWithToken, "with-token", false, "Authenticate with an API token")
	authLoginCmd.Flags().BoolVar(&flagWithOAuth, "with-oauth", false, "Authenticate with OAuth (browser)")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}
