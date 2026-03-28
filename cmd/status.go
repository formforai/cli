package cmd

import (
	"github.com/formforai/cli/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [form_id]",
	Short: "Check the status of a form",
	Long: `Display detailed information about a form.

Example:
  ff status form_abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		formID := args[0]
		c := getClient()

		detail, err := c.GetForm(formID)
		if err != nil {
			return err
		}

		if jsonOutput {
			output.PrintJSON(detail)
			return nil
		}

		output.FormDetail(
			detail.ID,
			detail.Title,
			detail.Status,
			detail.Recipient,
			detail.URL,
			detail.CreatedAt,
			detail.ExpiresAt,
		)

		if detail.Response != nil {
			cmd.Println()
			cmd.Println("Response:")
			output.PrintJSON(detail.Response.Data)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
