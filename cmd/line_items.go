package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nusii/nusii-cli/internal/models"
	"github.com/nusii/nusii-cli/internal/output"
)

var lineItemsCmd = &cobra.Command{
	Use:   "line-items",
	Short: "Manage line items",
	Long:  "List, create, update, and delete line items within sections.",
}

var lineItemsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List line items",
	Long: `List line items, optionally filtered by section.

Examples:
  nusii line-items list --section-id 123
  nusii line-items list --page 1 --per-page 10
  nusii line-items list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		sectionID, _ := cmd.Flags().GetString("section-id")

		raw, result, err := client.ListLineItems(page, perPage, sectionID)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.LineItemTableHeaders()
		var rows [][]string
		for _, r := range result.Data {
			rows = append(rows, r.Attributes.LineItemTableRow(r.ID))
		}
		output.PrintTable(headers, rows)
		if result.Meta != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Page %d of %d (%d total)\n", result.Meta.CurrentPage, result.Meta.TotalPages, result.Meta.TotalCount)
		}
		return nil
	},
}

var lineItemsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a line item by ID",
	Long: `Get details for a specific line item.

Examples:
  nusii line-items get 123
  nusii line-items get 123 -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		raw, result, err := client.GetLineItem(args[0])
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.LineItemTableHeaders()
		rows := [][]string{result.Data.Attributes.LineItemTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var lineItemsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new line item",
	Long: `Create a new line item under a section.

Examples:
  nusii line-items create --section-id 123 --name "Design" --amount 50000 --quantity 1
  nusii line-items create --section-id 123 --name "Development" --cost-type hourly --amount 15000 --quantity 40`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sectionID, _ := cmd.Flags().GetString("section-id")
		if sectionID == "" {
			return fmt.Errorf("--section-id is required")
		}

		attrs := models.LineItem{}
		attrs.Name, _ = cmd.Flags().GetString("name")
		attrs.CostType, _ = cmd.Flags().GetString("cost-type")
		attrs.RecurringType, _ = cmd.Flags().GetString("recurring-type")
		attrs.PerType, _ = cmd.Flags().GetString("per-type")
		quantity, _ := cmd.Flags().GetInt("quantity")
		attrs.Quantity = quantity
		amount, _ := cmd.Flags().GetInt("amount")
		attrs.AmountInCents = amount

		client := newAPIClient()
		raw, result, err := client.CreateLineItem(sectionID, attrs)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.LineItemTableHeaders()
		rows := [][]string{result.Data.Attributes.LineItemTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var lineItemsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an existing line item",
	Long: `Update a line item by ID.

Examples:
  nusii line-items update 123 --name "Updated Design" --amount 60000
  nusii line-items update 123 --quantity 50`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		attrs := models.LineItem{}
		if cmd.Flags().Changed("name") {
			attrs.Name, _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("cost-type") {
			attrs.CostType, _ = cmd.Flags().GetString("cost-type")
		}
		if cmd.Flags().Changed("recurring-type") {
			attrs.RecurringType, _ = cmd.Flags().GetString("recurring-type")
		}
		if cmd.Flags().Changed("per-type") {
			attrs.PerType, _ = cmd.Flags().GetString("per-type")
		}
		if cmd.Flags().Changed("quantity") {
			v, _ := cmd.Flags().GetInt("quantity")
			attrs.Quantity = v
		}
		if cmd.Flags().Changed("amount") {
			v, _ := cmd.Flags().GetInt("amount")
			attrs.AmountInCents = v
		}

		client := newAPIClient()
		raw, result, err := client.UpdateLineItem(args[0], attrs)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.LineItemTableHeaders()
		rows := [][]string{result.Data.Attributes.LineItemTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var lineItemsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a line item",
	Long: `Delete a line item by ID.

Examples:
  nusii line-items delete 123
  nusii line-items delete 123 --confirm`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		confirm, _ := cmd.Flags().GetBool("confirm")
		if !confirmAction("Delete line item "+args[0], confirm) {
			fmt.Println("Cancelled.")
			return nil
		}

		client := newAPIClient()
		if err := client.DeleteLineItem(args[0]); err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			fmt.Println(`{"status":"deleted"}`)
		} else {
			fmt.Printf("Line item %s deleted.\n", args[0])
		}
		return nil
	},
}

func init() {
	// List flags
	lineItemsListCmd.Flags().Int("page", 0, "Page number")
	lineItemsListCmd.Flags().Int("per-page", 0, "Items per page")
	lineItemsListCmd.Flags().String("section-id", "", "Filter by section ID")

	// Create flags
	lineItemsCreateCmd.Flags().String("section-id", "", "Section ID (required)")
	lineItemsCreateCmd.Flags().String("name", "", "Line item name")
	lineItemsCreateCmd.Flags().String("cost-type", "", "Cost type")
	lineItemsCreateCmd.Flags().String("recurring-type", "", "Recurring type")
	lineItemsCreateCmd.Flags().String("per-type", "", "Per type")
	lineItemsCreateCmd.Flags().Int("quantity", 0, "Quantity")
	lineItemsCreateCmd.Flags().Int("amount", 0, "Amount in cents")

	// Update flags
	lineItemsUpdateCmd.Flags().String("name", "", "Line item name")
	lineItemsUpdateCmd.Flags().String("cost-type", "", "Cost type")
	lineItemsUpdateCmd.Flags().String("recurring-type", "", "Recurring type")
	lineItemsUpdateCmd.Flags().String("per-type", "", "Per type")
	lineItemsUpdateCmd.Flags().Int("quantity", 0, "Quantity")
	lineItemsUpdateCmd.Flags().Int("amount", 0, "Amount in cents")

	// Delete flags
	lineItemsDeleteCmd.Flags().Bool("confirm", false, "Skip confirmation prompt")

	lineItemsCmd.AddCommand(lineItemsListCmd)
	lineItemsCmd.AddCommand(lineItemsGetCmd)
	lineItemsCmd.AddCommand(lineItemsCreateCmd)
	lineItemsCmd.AddCommand(lineItemsUpdateCmd)
	lineItemsCmd.AddCommand(lineItemsDeleteCmd)
	rootCmd.AddCommand(lineItemsCmd)
}
