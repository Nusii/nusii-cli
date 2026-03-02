package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nusii/nusii-cli/internal/models"
	"github.com/nusii/nusii-cli/internal/output"
)

var clientsCmd = &cobra.Command{
	Use:   "clients",
	Short: "Manage clients",
	Long:  "List, create, update, and delete clients.",
}

var clientsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all clients",
	Long: `List all clients with pagination support.

Examples:
  nusii clients list
  nusii clients list --page 2 --per-page 10
  nusii clients list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		raw, result, err := client.ListClients(page, perPage)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.ClientTableHeaders()
		var rows [][]string
		for _, r := range result.Data {
			rows = append(rows, r.Attributes.ClientTableRow(r.ID))
		}
		output.PrintTable(headers, rows)
		if result.Meta != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Page %d of %d (%d total)\n", result.Meta.CurrentPage, result.Meta.TotalPages, result.Meta.TotalCount)
		}
		return nil
	},
}

var clientsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a client by ID",
	Long: `Get details for a specific client.

Examples:
  nusii clients get 123
  nusii clients get 123 -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		raw, result, err := client.GetClient(args[0])
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.ClientTableHeaders()
		rows := [][]string{result.Data.Attributes.ClientTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var clientsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new client",
	Long: `Create a new client.

Examples:
  nusii clients create --name "John" --email "john@example.com"
  nusii clients create --name "John" --surname "Doe" --business "Acme Inc" -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		attrs := models.Client{}
		attrs.Name, _ = cmd.Flags().GetString("name")
		attrs.Email, _ = cmd.Flags().GetString("email")
		attrs.Surname, _ = cmd.Flags().GetString("surname")
		attrs.Business, _ = cmd.Flags().GetString("business")
		attrs.Currency, _ = cmd.Flags().GetString("currency")
		attrs.Locale, _ = cmd.Flags().GetString("locale")
		attrs.Web, _ = cmd.Flags().GetString("web")
		attrs.Telephone, _ = cmd.Flags().GetString("telephone")
		attrs.Address, _ = cmd.Flags().GetString("address")
		attrs.City, _ = cmd.Flags().GetString("city")
		attrs.Postcode, _ = cmd.Flags().GetString("postcode")
		attrs.Country, _ = cmd.Flags().GetString("country")
		attrs.State, _ = cmd.Flags().GetString("state")

		client := newAPIClient()
		raw, result, err := client.CreateClient(attrs)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.ClientTableHeaders()
		rows := [][]string{result.Data.Attributes.ClientTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var clientsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an existing client",
	Long: `Update a client by ID. Only specified flags are updated.

Examples:
  nusii clients update 123 --name "Jane"
  nusii clients update 123 --email "new@example.com" --business "New Corp"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		attrs := models.Client{}
		if cmd.Flags().Changed("name") {
			attrs.Name, _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("email") {
			attrs.Email, _ = cmd.Flags().GetString("email")
		}
		if cmd.Flags().Changed("surname") {
			attrs.Surname, _ = cmd.Flags().GetString("surname")
		}
		if cmd.Flags().Changed("business") {
			attrs.Business, _ = cmd.Flags().GetString("business")
		}
		if cmd.Flags().Changed("currency") {
			attrs.Currency, _ = cmd.Flags().GetString("currency")
		}
		if cmd.Flags().Changed("locale") {
			attrs.Locale, _ = cmd.Flags().GetString("locale")
		}
		if cmd.Flags().Changed("web") {
			attrs.Web, _ = cmd.Flags().GetString("web")
		}
		if cmd.Flags().Changed("telephone") {
			attrs.Telephone, _ = cmd.Flags().GetString("telephone")
		}
		if cmd.Flags().Changed("address") {
			attrs.Address, _ = cmd.Flags().GetString("address")
		}
		if cmd.Flags().Changed("city") {
			attrs.City, _ = cmd.Flags().GetString("city")
		}
		if cmd.Flags().Changed("postcode") {
			attrs.Postcode, _ = cmd.Flags().GetString("postcode")
		}
		if cmd.Flags().Changed("country") {
			attrs.Country, _ = cmd.Flags().GetString("country")
		}
		if cmd.Flags().Changed("state") {
			attrs.State, _ = cmd.Flags().GetString("state")
		}

		client := newAPIClient()
		raw, result, err := client.UpdateClient(args[0], attrs)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.ClientTableHeaders()
		rows := [][]string{result.Data.Attributes.ClientTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

var clientsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a client",
	Long: `Delete a client by ID.

Examples:
  nusii clients delete 123
  nusii clients delete 123 --confirm`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		confirm, _ := cmd.Flags().GetBool("confirm")
		if !confirmAction("Delete client "+args[0], confirm) {
			fmt.Println("Cancelled.")
			return nil
		}

		client := newAPIClient()
		if err := client.DeleteClient(args[0]); err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			fmt.Println(`{"status":"deleted"}`)
		} else {
			fmt.Printf("Client %s deleted.\n", args[0])
		}
		return nil
	},
}

func init() {
	// List flags
	clientsListCmd.Flags().Int("page", 0, "Page number")
	clientsListCmd.Flags().Int("per-page", 0, "Items per page")

	// Create flags
	clientsCreateCmd.Flags().String("name", "", "Client first name")
	clientsCreateCmd.Flags().String("email", "", "Client email")
	clientsCreateCmd.Flags().String("surname", "", "Client surname")
	clientsCreateCmd.Flags().String("business", "", "Business name")
	clientsCreateCmd.Flags().String("currency", "", "Currency code (e.g., USD, EUR)")
	clientsCreateCmd.Flags().String("locale", "", "Locale (e.g., en, nl)")
	clientsCreateCmd.Flags().String("web", "", "Website URL")
	clientsCreateCmd.Flags().String("telephone", "", "Phone number")
	clientsCreateCmd.Flags().String("address", "", "Street address")
	clientsCreateCmd.Flags().String("city", "", "City")
	clientsCreateCmd.Flags().String("postcode", "", "Postal code")
	clientsCreateCmd.Flags().String("country", "", "Country")
	clientsCreateCmd.Flags().String("state", "", "State/Province")

	// Update flags (same as create)
	clientsUpdateCmd.Flags().String("name", "", "Client first name")
	clientsUpdateCmd.Flags().String("email", "", "Client email")
	clientsUpdateCmd.Flags().String("surname", "", "Client surname")
	clientsUpdateCmd.Flags().String("business", "", "Business name")
	clientsUpdateCmd.Flags().String("currency", "", "Currency code")
	clientsUpdateCmd.Flags().String("locale", "", "Locale")
	clientsUpdateCmd.Flags().String("web", "", "Website URL")
	clientsUpdateCmd.Flags().String("telephone", "", "Phone number")
	clientsUpdateCmd.Flags().String("address", "", "Street address")
	clientsUpdateCmd.Flags().String("city", "", "City")
	clientsUpdateCmd.Flags().String("postcode", "", "Postal code")
	clientsUpdateCmd.Flags().String("country", "", "Country")
	clientsUpdateCmd.Flags().String("state", "", "State/Province")

	// Delete flags
	clientsDeleteCmd.Flags().Bool("confirm", false, "Skip confirmation prompt")

	clientsCmd.AddCommand(clientsListCmd)
	clientsCmd.AddCommand(clientsGetCmd)
	clientsCmd.AddCommand(clientsCreateCmd)
	clientsCmd.AddCommand(clientsUpdateCmd)
	clientsCmd.AddCommand(clientsDeleteCmd)
	rootCmd.AddCommand(clientsCmd)
}
