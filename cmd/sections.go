package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nusii/nusii-cli/internal/models"
	"github.com/nusii/nusii-cli/internal/output"
)

var sectionsCmd = &cobra.Command{
	Use:   "sections",
	Short: "Manage sections",
	Long:  "List, create, update, and delete proposal sections.",
}

var sectionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sections",
	Long: `List sections with optional filtering.

Examples:
  nusii sections list --proposal-id 123
  nusii sections list --template-id 456 --include-line-items
  nusii sections list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		proposalID, _ := cmd.Flags().GetString("proposal-id")
		templateID, _ := cmd.Flags().GetString("template-id")
		includeLineItems, _ := cmd.Flags().GetBool("include-line-items")

		raw, result, err := client.ListSections(page, perPage, proposalID, templateID, includeLineItems)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.SectionTableHeaders()
		var rows [][]string
		for _, r := range result.Data {
			rows = append(rows, r.Attributes.SectionTableRow(r.ID))
		}
		output.PrintTable(headers, rows)
		if result.Meta != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Page %d of %d (%d total)\n", result.Meta.CurrentPage, result.Meta.TotalPages, result.Meta.TotalCount)
		}
		return nil
	},
}

var sectionsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a section by ID",
	Long: `Get details for a specific section.

Examples:
  nusii sections get 123
  nusii sections get 123 -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		raw, result, err := client.GetSection(args[0])
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.SectionTableHeaders()
		rows := [][]string{result.Data.Attributes.SectionTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var sectionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new section",
	Long: `Create a new section on a proposal or template.

Examples:
  nusii sections create --proposal-id 123 --title "Pricing" --section-type cost
  nusii sections create --proposal-id 123 --title "Overview" --body "Project description..."`,
	RunE: func(cmd *cobra.Command, args []string) error {
		attrs := models.Section{}
		proposalID, _ := cmd.Flags().GetInt("proposal-id")
		attrs.ProposalID = proposalID
		templateID, _ := cmd.Flags().GetInt("template-id")
		attrs.TemplateID = templateID
		attrs.Title, _ = cmd.Flags().GetString("title")
		attrs.Name, _ = cmd.Flags().GetString("name")
		attrs.Body, _ = cmd.Flags().GetString("body")
		position, _ := cmd.Flags().GetInt("position")
		attrs.Position = position
		attrs.SectionType, _ = cmd.Flags().GetString("section-type")
		attrs.Reusable, _ = cmd.Flags().GetBool("reusable")
		attrs.Optional, _ = cmd.Flags().GetBool("optional")
		attrs.IncludeTotal, _ = cmd.Flags().GetBool("include-total")
		attrs.PageBreak, _ = cmd.Flags().GetBool("page-break")

		client := newAPIClient()
		raw, result, err := client.CreateSection(attrs)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.SectionTableHeaders()
		rows := [][]string{result.Data.Attributes.SectionTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var sectionsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an existing section",
	Long: `Update a section by ID.

Examples:
  nusii sections update 123 --title "Updated Pricing"
  nusii sections update 123 --body "New content"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		attrs := models.Section{}
		if cmd.Flags().Changed("proposal-id") {
			v, _ := cmd.Flags().GetInt("proposal-id")
			attrs.ProposalID = v
		}
		if cmd.Flags().Changed("template-id") {
			v, _ := cmd.Flags().GetInt("template-id")
			attrs.TemplateID = v
		}
		if cmd.Flags().Changed("title") {
			attrs.Title, _ = cmd.Flags().GetString("title")
		}
		if cmd.Flags().Changed("name") {
			attrs.Name, _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("body") {
			attrs.Body, _ = cmd.Flags().GetString("body")
		}
		if cmd.Flags().Changed("position") {
			v, _ := cmd.Flags().GetInt("position")
			attrs.Position = v
		}
		if cmd.Flags().Changed("section-type") {
			attrs.SectionType, _ = cmd.Flags().GetString("section-type")
		}
		if cmd.Flags().Changed("reusable") {
			attrs.Reusable, _ = cmd.Flags().GetBool("reusable")
		}
		if cmd.Flags().Changed("optional") {
			attrs.Optional, _ = cmd.Flags().GetBool("optional")
		}
		if cmd.Flags().Changed("include-total") {
			attrs.IncludeTotal, _ = cmd.Flags().GetBool("include-total")
		}
		if cmd.Flags().Changed("page-break") {
			attrs.PageBreak, _ = cmd.Flags().GetBool("page-break")
		}

		client := newAPIClient()
		raw, result, err := client.UpdateSection(args[0], attrs)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.SectionTableHeaders()
		rows := [][]string{result.Data.Attributes.SectionTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var sectionsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a section",
	Long: `Delete a section by ID.

Examples:
  nusii sections delete 123
  nusii sections delete 123 --confirm`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		confirm, _ := cmd.Flags().GetBool("confirm")
		if !confirmAction("Delete section "+args[0], confirm) {
			fmt.Println("Cancelled.")
			return nil
		}

		client := newAPIClient()
		if err := client.DeleteSection(args[0]); err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			fmt.Println(`{"status":"deleted"}`)
		} else {
			fmt.Printf("Section %s deleted.\n", args[0])
		}
		return nil
	},
}

func init() {
	// List flags
	sectionsListCmd.Flags().Int("page", 0, "Page number")
	sectionsListCmd.Flags().Int("per-page", 0, "Items per page")
	sectionsListCmd.Flags().String("proposal-id", "", "Filter by proposal ID")
	sectionsListCmd.Flags().String("template-id", "", "Filter by template ID")
	sectionsListCmd.Flags().Bool("include-line-items", false, "Include line items in response")

	// Create flags
	sectionsCreateCmd.Flags().Int("proposal-id", 0, "Proposal ID")
	sectionsCreateCmd.Flags().Int("template-id", 0, "Template ID")
	sectionsCreateCmd.Flags().String("title", "", "Section title")
	sectionsCreateCmd.Flags().String("name", "", "Section name")
	sectionsCreateCmd.Flags().String("body", "", "Section body content")
	sectionsCreateCmd.Flags().Int("position", 0, "Position in proposal")
	sectionsCreateCmd.Flags().String("section-type", "", "Section type (e.g., cost, text)")
	sectionsCreateCmd.Flags().Bool("reusable", false, "Make section reusable")
	sectionsCreateCmd.Flags().Bool("optional", false, "Make section optional")
	sectionsCreateCmd.Flags().Bool("include-total", false, "Include total in section")
	sectionsCreateCmd.Flags().Bool("page-break", false, "Add page break before section")

	// Update flags
	sectionsUpdateCmd.Flags().Int("proposal-id", 0, "Proposal ID")
	sectionsUpdateCmd.Flags().Int("template-id", 0, "Template ID")
	sectionsUpdateCmd.Flags().String("title", "", "Section title")
	sectionsUpdateCmd.Flags().String("name", "", "Section name")
	sectionsUpdateCmd.Flags().String("body", "", "Section body content")
	sectionsUpdateCmd.Flags().Int("position", 0, "Position")
	sectionsUpdateCmd.Flags().String("section-type", "", "Section type")
	sectionsUpdateCmd.Flags().Bool("reusable", false, "Reusable")
	sectionsUpdateCmd.Flags().Bool("optional", false, "Optional")
	sectionsUpdateCmd.Flags().Bool("include-total", false, "Include total")
	sectionsUpdateCmd.Flags().Bool("page-break", false, "Page break")

	// Delete flags
	sectionsDeleteCmd.Flags().Bool("confirm", false, "Skip confirmation prompt")

	sectionsCmd.AddCommand(sectionsListCmd)
	sectionsCmd.AddCommand(sectionsGetCmd)
	sectionsCmd.AddCommand(sectionsCreateCmd)
	sectionsCmd.AddCommand(sectionsUpdateCmd)
	sectionsCmd.AddCommand(sectionsDeleteCmd)
	rootCmd.AddCommand(sectionsCmd)
}
