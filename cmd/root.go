package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/nusii/nusii-cli/internal/api"
	"github.com/nusii/nusii-cli/internal/auth"
	"github.com/nusii/nusii-cli/internal/config"
	"github.com/nusii/nusii-cli/internal/output"
)

var (
	flagAPIKey  string
	flagAPIURL  string
	flagOutput  string
	flagNoInput bool
	flagDebug   bool

	cfg *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "nusii",
	Short: "Nusii CLI - manage proposals from the command line",
	Long:  "A command-line interface for the Nusii proposal software API.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Flag overrides
		if flagAPIKey != "" {
			cfg.APIKey = flagAPIKey
		}
		if flagAPIURL != "" {
			cfg.APIURL = flagAPIURL
		}
		if flagOutput != "" {
			cfg.Output = flagOutput
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagAPIKey, "api-key", "k", "", "API key (overrides config and NUSII_API_KEY)")
	rootCmd.PersistentFlags().StringVar(&flagAPIURL, "api-url", "", "API base URL (overrides config and NUSII_API_URL)")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format: json or table")
	rootCmd.PersistentFlags().BoolVar(&flagNoInput, "no-input", false, "Disable interactive prompts")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Print HTTP request/response details to stderr")
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		// Check for API error to set exit code
		if apiErr, ok := err.(*api.APIError); ok {
			format := getOutputFormat()
			if format == output.FormatJSON {
				output.PrintErrorJSON(apiErr.Message, apiErr.StatusCode)
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s\n", apiErr.Message)
			}
			os.Exit(apiErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return err
	}
	return nil
}

// newAPIClient creates an authenticated API client from the current config.
func newAPIClient() *api.Client {
	var authenticator auth.Authenticator

	if flagAPIKey != "" || (cfg.APIKey != "" && !cfg.HasOAuthToken()) {
		authenticator = auth.NewTokenAuth(cfg.APIKey)
	} else if cfg.HasOAuthToken() {
		expiry, _ := cfg.OAuthExpiryTime()
		clientID := resolveOAuthClientID()
		tokenURL := cfg.APIURL + "/oauth/token"
		authenticator = &auth.OAuthAuth{
			AccessToken:  cfg.OAuthAccessToken,
			RefreshToken: cfg.OAuthRefreshToken,
			Expiry:       expiry,
			TokenURL:     tokenURL,
			ClientID:     clientID,
			OnRefresh: func(accessToken, refreshToken string, newExpiry time.Time) error {
				return config.SaveOAuthTokens(accessToken, refreshToken, newExpiry, clientID)
			},
		}
	} else {
		authenticator = auth.NewTokenAuth("")
	}

	client := api.NewClient(cfg.APIURL, authenticator)
	client.Debug = flagDebug
	client.Version = version
	return client
}

// getOutputFormat returns the resolved output format.
func getOutputFormat() output.Format {
	return output.Detect(cfg.Output)
}

// confirmAction prompts the user for confirmation unless --no-input or --confirm is set.
func confirmAction(action string, confirmed bool) bool {
	if confirmed || flagNoInput {
		return true
	}
	fmt.Fprintf(os.Stderr, "%s? [y/N] ", action)
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes"
}
