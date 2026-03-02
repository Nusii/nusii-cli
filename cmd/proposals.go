package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nusii/nusii-cli/internal/models"
	"github.com/nusii/nusii-cli/internal/output"
)

var proposalsCmd = &cobra.Command{
	Use:   "proposals",
	Short: "Manage proposals",
	Long:  "List, create, update, delete, send, and archive proposals.",
}

var proposalsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List proposals",
	Long: `List proposals with optional filtering.

Examples:
  nusii proposals list
  nusii proposals list --status draft
  nusii proposals list --archived --page 1 --per-page 5
  nusii proposals list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		status, _ := cmd.Flags().GetString("status")
		archived, _ := cmd.Flags().GetBool("archived")

		raw, result, err := client.ListProposals(page, perPage, status, archived)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.ProposalTableHeaders()
		var rows [][]string
		for _, r := range result.Data {
			rows = append(rows, r.Attributes.ProposalTableRow(r.ID))
		}
		output.PrintTable(headers, rows)
		if result.Meta != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Page %d of %d (%d total)\n", result.Meta.CurrentPage, result.Meta.TotalPages, result.Meta.TotalCount)
		}
		return nil
	},
}

var proposalsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a proposal by ID",
	Long: `Get details for a specific proposal.

Examples:
  nusii proposals get 123
  nusii proposals get 123 -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		raw, result, err := client.GetProposal(args[0])
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.ProposalTableHeaders()
		rows := [][]string{result.Data.Attributes.ProposalTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var proposalsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new proposal",
	Long: `Create a new proposal.

Examples:
  nusii proposals create --title "Web Design" --client-id 123
  nusii proposals create --title "Web Design" --client-email "john@example.com" -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		attrs := models.Proposal{}
		attrs.Title, _ = cmd.Flags().GetString("title")
		clientID, _ := cmd.Flags().GetInt("client-id")
		attrs.ClientID = clientID
		attrs.ClientEmail, _ = cmd.Flags().GetString("client-email")
		templateID, _ := cmd.Flags().GetInt("template-id")
		attrs.TemplateID = templateID
		attrs.Theme, _ = cmd.Flags().GetString("theme")
		attrs.Currency, _ = cmd.Flags().GetString("currency")
		attrs.ExpiresAt, _ = cmd.Flags().GetString("expires-at")
		attrs.DisplayDate, _ = cmd.Flags().GetString("display-date")
		attrs.Report, _ = cmd.Flags().GetBool("report")
		attrs.ExcludeTotal, _ = cmd.Flags().GetBool("exclude-total")

		client := newAPIClient()
		raw, result, err := client.CreateProposal(attrs)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.ProposalTableHeaders()
		rows := [][]string{result.Data.Attributes.ProposalTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var proposalsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an existing proposal",
	Long: `Update a proposal by ID.

Examples:
  nusii proposals update 123 --title "New Title"
  nusii proposals update 123 --currency EUR`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		attrs := models.Proposal{}
		if cmd.Flags().Changed("title") {
			attrs.Title, _ = cmd.Flags().GetString("title")
		}
		if cmd.Flags().Changed("client-id") {
			v, _ := cmd.Flags().GetInt("client-id")
			attrs.ClientID = v
		}
		if cmd.Flags().Changed("client-email") {
			attrs.ClientEmail, _ = cmd.Flags().GetString("client-email")
		}
		if cmd.Flags().Changed("template-id") {
			v, _ := cmd.Flags().GetInt("template-id")
			attrs.TemplateID = v
		}
		if cmd.Flags().Changed("theme") {
			attrs.Theme, _ = cmd.Flags().GetString("theme")
		}
		if cmd.Flags().Changed("currency") {
			attrs.Currency, _ = cmd.Flags().GetString("currency")
		}
		if cmd.Flags().Changed("expires-at") {
			attrs.ExpiresAt, _ = cmd.Flags().GetString("expires-at")
		}
		if cmd.Flags().Changed("display-date") {
			attrs.DisplayDate, _ = cmd.Flags().GetString("display-date")
		}
		if cmd.Flags().Changed("report") {
			attrs.Report, _ = cmd.Flags().GetBool("report")
		}
		if cmd.Flags().Changed("exclude-total") {
			attrs.ExcludeTotal, _ = cmd.Flags().GetBool("exclude-total")
		}

		client := newAPIClient()
		raw, result, err := client.UpdateProposal(args[0], attrs)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.ProposalTableHeaders()
		rows := [][]string{result.Data.Attributes.ProposalTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var proposalsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a proposal",
	Long: `Delete a proposal by ID.

Examples:
  nusii proposals delete 123
  nusii proposals delete 123 --confirm`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		confirm, _ := cmd.Flags().GetBool("confirm")
		if !confirmAction("Delete proposal "+args[0], confirm) {
			fmt.Println("Cancelled.")
			return nil
		}

		client := newAPIClient()
		if err := client.DeleteProposal(args[0]); err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			fmt.Println(`{"status":"deleted"}`)
		} else {
			fmt.Printf("Proposal %s deleted.\n", args[0])
		}
		return nil
	},
}

var proposalsSendCmd = &cobra.Command{
	Use:   "send <id>",
	Short: "Send a proposal",
	Long: `Send a proposal to the client.

Examples:
  nusii proposals send 123
  nusii proposals send 123 --email "override@example.com" --subject "Your proposal"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sendReq := models.ProposalSendRequest{}
		sendReq.Email, _ = cmd.Flags().GetString("email")
		sendReq.CC, _ = cmd.Flags().GetString("cc")
		sendReq.BCC, _ = cmd.Flags().GetString("bcc")
		sendReq.Subject, _ = cmd.Flags().GetString("subject")
		sendReq.Message, _ = cmd.Flags().GetString("message")

		client := newAPIClient()
		raw, err := client.SendProposal(args[0], sendReq)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		fmt.Printf("Proposal %s sent.\n", args[0])
		return nil
	},
}

var proposalsArchiveCmd = &cobra.Command{
	Use:   "archive <id>",
	Short: "Archive a proposal",
	Long: `Archive a proposal by ID.

Examples:
  nusii proposals archive 123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		raw, err := client.ArchiveProposal(args[0])
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		fmt.Printf("Proposal %s archived.\n", args[0])
		_ = raw
		return nil
	},
}

func init() {
	// List flags
	proposalsListCmd.Flags().Int("page", 0, "Page number")
	proposalsListCmd.Flags().Int("per-page", 0, "Items per page")
	proposalsListCmd.Flags().String("status", "", "Filter by status (draft, pending, accepted, rejected)")
	proposalsListCmd.Flags().Bool("archived", false, "Include archived proposals")

	// Create flags
	proposalsCreateCmd.Flags().String("title", "", "Proposal title")
	proposalsCreateCmd.Flags().Int("client-id", 0, "Client ID")
	proposalsCreateCmd.Flags().String("client-email", "", "Client email (alternative to client-id)")
	proposalsCreateCmd.Flags().Int("template-id", 0, "Template ID to base proposal on")
	proposalsCreateCmd.Flags().String("theme", "", "Theme name")
	proposalsCreateCmd.Flags().String("currency", "", "Currency code")
	proposalsCreateCmd.Flags().String("expires-at", "", "Expiration date (YYYY-MM-DD)")
	proposalsCreateCmd.Flags().String("display-date", "", "Display date (YYYY-MM-DD)")
	proposalsCreateCmd.Flags().Bool("report", false, "Create as report (no pricing)")
	proposalsCreateCmd.Flags().Bool("exclude-total", false, "Exclude total from proposal")

	// Update flags
	proposalsUpdateCmd.Flags().String("title", "", "Proposal title")
	proposalsUpdateCmd.Flags().Int("client-id", 0, "Client ID")
	proposalsUpdateCmd.Flags().String("client-email", "", "Client email")
	proposalsUpdateCmd.Flags().Int("template-id", 0, "Template ID")
	proposalsUpdateCmd.Flags().String("theme", "", "Theme name")
	proposalsUpdateCmd.Flags().String("currency", "", "Currency code")
	proposalsUpdateCmd.Flags().String("expires-at", "", "Expiration date")
	proposalsUpdateCmd.Flags().String("display-date", "", "Display date")
	proposalsUpdateCmd.Flags().Bool("report", false, "Report mode")
	proposalsUpdateCmd.Flags().Bool("exclude-total", false, "Exclude total")

	// Delete flags
	proposalsDeleteCmd.Flags().Bool("confirm", false, "Skip confirmation prompt")

	// Send flags
	proposalsSendCmd.Flags().String("email", "", "Override recipient email")
	proposalsSendCmd.Flags().String("cc", "", "CC email address")
	proposalsSendCmd.Flags().String("bcc", "", "BCC email address")
	proposalsSendCmd.Flags().String("subject", "", "Email subject")
	proposalsSendCmd.Flags().String("message", "", "Email message body")

	proposalsCmd.AddCommand(proposalsListCmd)
	proposalsCmd.AddCommand(proposalsGetCmd)
	proposalsCmd.AddCommand(proposalsCreateCmd)
	proposalsCmd.AddCommand(proposalsUpdateCmd)
	proposalsCmd.AddCommand(proposalsDeleteCmd)
	proposalsCmd.AddCommand(proposalsSendCmd)
	proposalsCmd.AddCommand(proposalsArchiveCmd)
	rootCmd.AddCommand(proposalsCmd)
}
