package cmd

import (
	"github.com/formforai/cli/internal/output"
	"github.com/formforai/cli/internal/client"
	"github.com/spf13/cobra"
)

var (
	listStatus string
	listLimit  int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List forms",
	Long: `List forms with optional filters.

Examples:
  ff list
  ff list --status pending
  ff list --limit 5`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient()

		result, err := c.ListForms(client.ListOptions{
			Status: listStatus,
			Limit:  listLimit,
		})
		if err != nil {
			return err
		}

		if jsonOutput {
			output.PrintJSON(result)
			return nil
		}

		if len(result.Forms) == 0 {
			cmd.Println("No forms found.")
			return nil
		}

		headers := []string{"ID", "TITLE", "STATUS", "RECIPIENT", "CREATED"}
		rows := make([][]string, len(result.Forms))
		for i, f := range result.Forms {
			rows[i] = []string{
				f.ID,
				f.Title,
				output.StatusColor(f.Status),
				f.Recipient,
				output.RelativeTime(f.CreatedAt),
			}
		}
		output.Table(headers, rows)

		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (pending, completed, expired, cancelled)")
	listCmd.Flags().IntVar(&listLimit, "limit", 0, "Maximum number of forms to return")

	rootCmd.AddCommand(listCmd)
}
