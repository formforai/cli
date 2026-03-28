package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/formforai/cli/internal/client"
	"github.com/formforai/cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	collectTitle   string
	collectFields  []string
	collectSchema  string
	collectTo      string
	collectExpires string
	collectContext string
	collectWait    bool
	collectTimeout time.Duration
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect structured input via a form",
	Long: `Create a form with typed fields and send it to a recipient.

Field format:
  --field "id:type:label:options"

  id       Field identifier
  type     One of: text, email, select, textarea, number, boolean
  label    Display label
  options  Comma-separated options (for select) or "optional" to mark not required

Examples:
  ff collect --title "Bug triage" \
    --field "severity:select:Severity:P0,P1,P2,P3" \
    --field "assignee:email:Assign to" \
    --field "notes:textarea:Notes:optional" \
    --to eng-lead@company.com --wait

  ff collect --schema intake.json --to user@company.com --wait`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient()

		var fields []client.Field
		var err error

		if collectSchema != "" {
			fields, err = loadSchemaFile(collectSchema)
			if err != nil {
				return fmt.Errorf("loading schema file: %w", err)
			}
		}

		for _, raw := range collectFields {
			f, parseErr := parseField(raw)
			if parseErr != nil {
				return fmt.Errorf("invalid --field %q: %w", raw, parseErr)
			}
			fields = append(fields, f)
		}

		if len(fields) == 0 {
			return fmt.Errorf("no fields specified; use --field or --schema")
		}

		params := client.CreateFormParams{
			Title:     collectTitle,
			Fields:    fields,
			Recipient: collectTo,
			ExpiresIn: collectExpires,
			Context:   collectContext,
		}

		form, err := c.CreateForm(params)
		if err != nil {
			return err
		}

		if jsonOutput && !collectWait {
			output.PrintJSON(form)
			return nil
		}

		output.FormCreated(form.ID, form.URL)

		if collectWait {
			output.Waiting()
			resp, err := c.WaitForResponse(form.ID, collectTimeout)
			if err != nil {
				return err
			}
			output.PrintJSON(resp.Data)
		}

		return nil
	},
}

// parseField parses "id:type:label:options" into a Field.
func parseField(raw string) (client.Field, error) {
	parts := strings.SplitN(raw, ":", 4)
	if len(parts) < 3 {
		return client.Field{}, fmt.Errorf("expected at least id:type:label, got %d parts", len(parts))
	}

	f := client.Field{
		ID:       parts[0],
		Type:     parts[1],
		Label:    parts[2],
		Required: true,
	}

	if len(parts) == 4 {
		opts := parts[3]
		if opts == "optional" {
			f.Required = false
		} else if f.Type == "select" {
			f.Options = strings.Split(opts, ",")
		} else {
			// Could be "optional" for other types or options for select.
			if opts == "optional" {
				f.Required = false
			}
		}
	}

	return f, nil
}

// schemaField is the JSON representation used in schema files.
type schemaField struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Label    string   `json:"label"`
	Required *bool    `json:"required,omitempty"`
	Options  []string `json:"options,omitempty"`
}

type schemaFile struct {
	Title  string        `json:"title,omitempty"`
	Fields []schemaField `json:"fields"`
}

func loadSchemaFile(path string) ([]client.Field, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s schemaFile
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	if s.Title != "" && collectTitle == "" {
		collectTitle = s.Title
	}

	fields := make([]client.Field, len(s.Fields))
	for i, sf := range s.Fields {
		required := true
		if sf.Required != nil {
			required = *sf.Required
		}
		fields[i] = client.Field{
			ID:       sf.ID,
			Type:     sf.Type,
			Label:    sf.Label,
			Required: required,
			Options:  sf.Options,
		}
	}
	return fields, nil
}

func init() {
	collectCmd.Flags().StringVar(&collectTitle, "title", "", "Form title")
	collectCmd.Flags().StringArrayVar(&collectFields, "field", nil, "Field definition (id:type:label:options)")
	collectCmd.Flags().StringVar(&collectSchema, "schema", "", "Path to a JSON schema file defining fields")
	collectCmd.Flags().StringVar(&collectTo, "to", "", "Recipient email address (required)")
	collectCmd.Flags().StringVar(&collectExpires, "expires", "", "Expiration duration (e.g. 4h, 30m, 1d)")
	collectCmd.Flags().StringVar(&collectContext, "context", "", "Additional context shown on the form")
	collectCmd.Flags().BoolVar(&collectWait, "wait", false, "Wait for a response before exiting")
	collectCmd.Flags().DurationVar(&collectTimeout, "timeout", 24*time.Hour, "Maximum time to wait for a response")

	_ = collectCmd.MarkFlagRequired("to")

	rootCmd.AddCommand(collectCmd)
}
