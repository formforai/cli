package cmd

import (
	"fmt"
	"os"

	"github.com/formforai/cli/internal/client"
	"github.com/formforai/cli/internal/config"
	"github.com/formforai/cli/internal/output"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

// Global flags.
var (
	apiKeyFlag string
	apiURLFlag string
	jsonOutput bool
)

var rootCmd = &cobra.Command{
	Use:     "ff",
	Short:   "FormFor CLI -- structured input for AI agents",
	Long:    "Create and manage forms that collect human input for AI agents.",
	Version: version,
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "API key (overrides config and FF_API_KEY)")
	rootCmd.PersistentFlags().StringVar(&apiURLFlag, "api-url", "", "API base URL (overrides config and FF_API_URL)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output raw JSON")
}

// getClient builds an API client from flags, env, and config (in that priority order).
func getClient() *client.Client {
	key := apiKeyFlag
	if key == "" {
		key = config.GetAPIKey()
	}
	if key == "" {
		output.Error("API key not configured. Set it with:\n  ff config set api-key <your-key>\nor export FF_API_KEY=<your-key>")
		os.Exit(1)
	}

	base := apiURLFlag
	if base == "" {
		base = config.GetAPIURL()
	}

	return client.New(key, base)
}

// exitErr prints an error and exits.
func exitErr(err error) {
	output.Error(fmt.Sprintf("%v", err))
	os.Exit(1)
}
