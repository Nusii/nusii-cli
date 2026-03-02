package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nusii/nusii-cli/internal/models"
	"github.com/nusii/nusii-cli/internal/output"
)

var activitiesCmd = &cobra.Command{
	Use:   "activities",
	Short: "View proposal activities",
	Long:  "List and view proposal activity events.",
}

var activitiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List activities",
	Long: `List proposal activities with optional filtering.

Examples:
  nusii activities list
  nusii activities list --proposal-id 123
  nusii activities list --client-id 456
  nusii activities list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		proposalID, _ := cmd.Flags().GetString("proposal-id")
		clientID, _ := cmd.Flags().GetString("client-id")

		raw, result, err := client.ListActivities(page, perPage, proposalID, clientID)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.ActivityTableHeaders()
		var rows [][]string
		for _, r := range result.Data {
			rows = append(rows, r.Attributes.ActivityTableRow(r.ID))
		}
		output.PrintTable(headers, rows)
		if result.Meta != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Page %d of %d (%d total)\n", result.Meta.CurrentPage, result.Meta.TotalPages, result.Meta.TotalCount)
		}
		return nil
	},
}

var activitiesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get an activity by ID",
	Long: `Get details for a specific activity.

Examples:
  nusii activities get 123
  nusii activities get 123 -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newAPIClient()
		raw, result, err := client.GetActivity(args[0])
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == output.FormatJSON {
			return output.PrintJSON(raw)
		}

		headers := models.ActivityTableHeaders()
		rows := [][]string{result.Data.Attributes.ActivityTableRow(result.Data.ID)}
		output.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	activitiesListCmd.Flags().Int("page", 0, "Page number")
	activitiesListCmd.Flags().Int("per-page", 0, "Items per page")
	activitiesListCmd.Flags().String("proposal-id", "", "Filter by proposal ID")
	activitiesListCmd.Flags().String("client-id", "", "Filter by client ID")

	activitiesCmd.AddCommand(activitiesListCmd)
	activitiesCmd.AddCommand(activitiesGetCmd)
	rootCmd.AddCommand(activitiesCmd)
}
