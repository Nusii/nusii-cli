package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nusii/nusii-cli/internal/models"
	"github.com/nusii/nusii-cli/internal/output"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Show account information",
	Long: `Display the current account details.

Examples:
  nusii account
  nusii account -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		raw, result, err := client.GetAccount()
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.AccountTableHeaders()
		rows := [][]string{result.Data.Attributes.AccountTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)
}
