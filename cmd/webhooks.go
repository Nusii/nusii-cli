package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nusii/nusii-cli/internal/models"
	"github.com/nusii/nusii-cli/internal/output"
)

var webhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage webhook endpoints",
	Long:  "List, create, and delete webhook endpoints.",
}

var webhooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhook endpoints",
	Long: `List all webhook endpoints.

Examples:
  nusii webhooks list
  nusii webhooks list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		raw, result, err := client.ListWebhooks(page, perPage)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.WebhookTableHeaders()
		var rows [][]string
		for _, r := range result.Data {
			rows = append(rows, r.Attributes.WebhookTableRow(r.ID))
		}
		output.PrintTable(headers, rows)
		if result.Meta != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Page %d of %d (%d total)\n", result.Meta.CurrentPage, result.Meta.TotalPages, result.Meta.TotalCount)
		}
		return nil
	},
}

var webhooksGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a webhook endpoint by ID",
	Long: `Get details for a specific webhook endpoint.

Examples:
  nusii webhooks get 123
  nusii webhooks get 123 -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		raw, result, err := client.GetWebhook(args[0])
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.WebhookTableHeaders()
		rows := [][]string{result.Data.Attributes.WebhookTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var webhooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new webhook endpoint",
	Long: `Create a new webhook endpoint.

Examples:
  nusii webhooks create --target-url "https://example.com/webhook" --events "proposal.sent,proposal.accepted"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetURL, _ := cmd.Flags().GetString("target-url")
		eventsStr, _ := cmd.Flags().GetString("events")

		if targetURL == "" {
			return fmt.Errorf("--target-url is required")
		}

		var events []string
		if eventsStr != "" {
			events = strings.Split(eventsStr, ",")
			for i := range events {
				events[i] = strings.TrimSpace(events[i])
			}
		}

		attrs := models.WebhookEndpoint{
			TargetURL: targetURL,
			Events:    events,
		}

		client := newAPIClient()
		raw, result, err := client.CreateWebhook(attrs)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.WebhookTableHeaders()
		rows := [][]string{result.Data.Attributes.WebhookTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var webhooksDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a webhook endpoint",
	Long: `Delete a webhook endpoint by ID.

Examples:
  nusii webhooks delete 123
  nusii webhooks delete 123 --confirm`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		confirm, _ := cmd.Flags().GetBool("confirm")
		if !confirmAction("Delete webhook "+args[0], confirm) {
			fmt.Println("Cancelled.")
			return nil
		}

		client := newAPIClient()
		if err := client.DeleteWebhook(args[0]); err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			fmt.Println(`{"status":"deleted"}`)
		} else {
			fmt.Printf("Webhook %s deleted.\n", args[0])
		}
		return nil
	},
}

func init() {
	// List flags
	webhooksListCmd.Flags().Int("page", 0, "Page number")
	webhooksListCmd.Flags().Int("per-page", 0, "Items per page")

	// Create flags
	webhooksCreateCmd.Flags().String("target-url", "", "Webhook target URL (required)")
	webhooksCreateCmd.Flags().String("events", "", "Comma-separated list of events")

	// Delete flags
	webhooksDeleteCmd.Flags().Bool("confirm", false, "Skip confirmation prompt")

	webhooksCmd.AddCommand(webhooksListCmd)
	webhooksCmd.AddCommand(webhooksGetCmd)
	webhooksCmd.AddCommand(webhooksCreateCmd)
	webhooksCmd.AddCommand(webhooksDeleteCmd)
	rootCmd.AddCommand(webhooksCmd)
}
