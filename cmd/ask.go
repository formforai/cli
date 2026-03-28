package cmd

import (
	"time"

	"github.com/formforai/cli/internal/client"
	"github.com/formforai/cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	askTo      string
	askContext string
	askExpires string
	askWait    bool
	askTimeout time.Duration
)

var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask a human a yes/no question",
	Long: `Send a yes/no approval question to a recipient and optionally
wait for their response.

Examples:
  ff ask "Deploy to production?" --to cto@company.com --wait
  ff ask "Approve refund?" --to finance@company.com --expires 4h --wait`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		question := args[0]
		c := getClient()

		form, err := c.Ask(question, askTo, client.AskOptions{
			Context: askContext,
			Expires: askExpires,
		})
		if err != nil {
			return err
		}

		if jsonOutput && !askWait {
			output.PrintJSON(form)
			return nil
		}

		output.FormCreated(form.ID, form.URL)

		if askWait {
			output.Waiting()
			resp, err := c.WaitForResponse(form.ID, askTimeout)
			if err != nil {
				return err
			}
			output.PrintJSON(resp.Data)
		}

		return nil
	},
}

func init() {
	askCmd.Flags().StringVar(&askTo, "to", "", "Recipient email address (required)")
	askCmd.Flags().StringVar(&askContext, "context", "", "Additional context shown on the form")
	askCmd.Flags().StringVar(&askExpires, "expires", "", "Expiration duration (e.g. 4h, 30m, 1d)")
	askCmd.Flags().BoolVar(&askWait, "wait", false, "Wait for a response before exiting")
	askCmd.Flags().DurationVar(&askTimeout, "timeout", 24*time.Hour, "Maximum time to wait for a response")

	_ = askCmd.MarkFlagRequired("to")

	rootCmd.AddCommand(askCmd)
}
