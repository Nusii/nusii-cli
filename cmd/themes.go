package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nusii/nusii-cli/internal/output"
)

var themesCmd = &cobra.Command{
	Use:   "themes",
	Short: "Manage themes",
	Long:  "List available proposal themes.",
}

var themesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List themes",
	Long: `List all available themes.

Examples:
  nusii themes list
  nusii themes list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		raw, result, err := client.ListThemes()
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := []string{"ID", "Name"}
		var rows [][]string
		for _, t := range result {
			rows = append(rows, []string{t.ID, t.Name})
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	themesCmd.AddCommand(themesListCmd)
	rootCmd.AddCommand(themesCmd)
}
