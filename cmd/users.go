package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nusii/nusii-cli/internal/models"
	"github.com/nusii/nusii-cli/internal/output"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
	Long:  "List users on the account.",
}

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List users",
	Long: `List all users on the account.

Examples:
  nusii users list
  nusii users list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		raw, result, err := client.ListUsers(page, perPage)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.UserTableHeaders()
		var rows [][]string
		for _, r := range result.Data {
			rows = append(rows, r.Attributes.UserTableRow(r.ID))
		}
		output.PrintTable(headers, rows)
		if result.Meta != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Page %d of %d (%d total)\n", result.Meta.CurrentPage, result.Meta.TotalPages, result.Meta.TotalCount)
		}
		return nil
	},
}

func init() {
	usersListCmd.Flags().Int("page", 0, "Page number")
	usersListCmd.Flags().Int("per-page", 0, "Items per page")

	usersCmd.AddCommand(usersListCmd)
	rootCmd.AddCommand(usersCmd)
}
