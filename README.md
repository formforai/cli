# FormFor CLI

Command-line interface for FormFor. Built in Go with Cobra. Create forms, collect structured input, and check responses from your terminal.

## Installation

```bash
go install github.com/formforai/cli@latest
```

The binary is named `ff`.

## Quick Start

```bash
# Configure your API key
ff config set api-key ff_live_...

# Ask a yes/no question
ff ask "Deploy to production?" --to ops@company.com --wait

# Collect structured data
ff collect --title "Bug Triage" \
  --field "severity:select:Severity:P0,P1,P2,P3" \
  --field "assignee:email:Assign to" \
  --field "notes:textarea:Notes:optional" \
  --to eng@company.com --wait

# List recent forms
ff list --status pending

# Check a specific form
ff status form_abc123
```

## Commands

### `ff ask [question]`

Send a yes/no question to a recipient.

```bash
ff ask "Approve refund?" --to finance@company.com --expires 4h --wait
```

| Flag | Description |
|------|-------------|
| `--to` | Recipient email (required) |
| `--context` | Additional context shown on the form |
| `--expires` | Expiry duration (e.g. `4h`, `30m`, `1d`) |
| `--wait` | Wait for response before exiting |
| `--timeout` | Max wait duration (default `24h`) |

### `ff collect`

Create a multi-field form.

```bash
ff collect --title "Intake Form" \
  --field "name:text:Full Name" \
  --field "role:select:Role:engineer,manager,other" \
  --schema intake.json \
  --to user@company.com --wait
```

Field format: `id:type:label:options`

- `options` can be comma-separated values (for `select`) or `optional` to mark not required

| Flag | Description |
|------|-------------|
| `--title` | Form title |
| `--field` | Field definition (repeatable) |
| `--schema` | Path to JSON schema file |
| `--to` | Recipient email (required) |
| `--context` | Additional context |
| `--expires` | Expiry duration |
| `--wait` | Wait for response |
| `--timeout` | Max wait duration (default `24h`) |

### `ff list`

List recent forms.

```bash
ff list --status pending --limit 20
```

### `ff status [form_id]`

Check the status of a specific form.

```bash
ff status form_abc123
```

### `ff config`

Manage CLI configuration.

```bash
ff config set api-key ff_live_...
ff config set api-url https://api.formfor.ai
ff config get api-key
```

## Configuration

The CLI resolves configuration in this order:

1. `--api-key` / `--api-url` flags
2. `FF_API_KEY` / `FF_API_URL` environment variables
3. Config file (set via `ff config set`)

## Global Flags

| Flag | Description |
|------|-------------|
| `--api-key` | API key (overrides config and env) |
| `--api-url` | API base URL (overrides config and env) |
| `--json` | Output raw JSON |
| `--version` | Print version |

## Development

```bash
go build -o ff .
./ff --version
```

## License

MIT
