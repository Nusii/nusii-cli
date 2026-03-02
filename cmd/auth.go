package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nusii/nusii-cli/internal/config"
	"github.com/nusii/nusii-cli/internal/output"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  "Login, check status, or logout from the Nusii API.",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Nusii API",
	Long: `Authenticate with the Nusii API using an API key.

Examples:
  nusii auth login
  nusii auth login --api-key YOUR_KEY
  nusii auth login --api-key YOUR_KEY --api-url http://localhost:3000`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		// Save to config
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
			fmt.Printf(`{"status":"authenticated","account":"%s","email":"%s"}`, account.Data.Attributes.Name, account.Data.Attributes.Email)
			fmt.Println()
		} else {
			fmt.Fprintf(os.Stderr, "Authenticated as %s (%s)\n", account.Data.Attributes.Name, account.Data.Attributes.Email)
			fmt.Fprintf(os.Stderr, "Config saved to %s\n", config.ConfigPath())
		}
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	Long: `Show current authentication status and account info.

Examples:
  nusii auth status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.APIKey == "" {
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
			fmt.Printf(`{"status":"authenticated","account":"%s","email":"%s","api_url":"%s"}`,
				account.Data.Attributes.Name, account.Data.Attributes.Email, cfg.APIURL)
			fmt.Println()
		} else {
			fmt.Printf("Authenticated as %s (%s)\n", account.Data.Attributes.Name, account.Data.Attributes.Email)
			fmt.Printf("API URL: %s\n", cfg.APIURL)
		}
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored authentication",
	Long: `Remove the stored API key from the config file.

Examples:
  nusii auth logout`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.RemoveAPIKey(); err != nil {
			return fmt.Errorf("removing API key: %w", err)
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			fmt.Println(`{"status":"logged out"}`)
		} else {
			fmt.Println("Logged out. API key removed from config.")
		}
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}
