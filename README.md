# Nusii CLI

A command-line interface for the [Nusii](https://nusii.com) proposal software API.

## Installation

### Homebrew

```bash
brew install nusii/tap/nusii
```

### From source

```bash
go install github.com/nusii/nusii-cli@latest
```

### Download binary

Download the latest release from the [releases page](https://github.com/Nusii/nusii-cli/releases).

## Authentication

```bash
nusii auth login --api-key YOUR_API_KEY
```

This saves your API key to `~/.config/nusii/config.yaml`. You can also set it via the `NUSII_API_KEY` environment variable.

```bash
# Check auth status
nusii auth status

# Remove stored credentials
nusii auth logout
```

## Usage

### Account

```bash
nusii account
```

### Clients

```bash
nusii clients list
nusii clients get 123
nusii clients create --name "John" --email "john@example.com" --business "Acme Inc"
nusii clients update 123 --name "Jane"
nusii clients delete 123
```

### Proposals

```bash
nusii proposals list
nusii proposals list --status accepted
nusii proposals get 123
nusii proposals create --title "Web Design" --client-id 456
nusii proposals update 123 --title "Updated Title"
nusii proposals send 123
nusii proposals archive 123
nusii proposals delete 123
```

### Sections

```bash
nusii sections list --proposal-id 123
nusii sections get 456
nusii sections create --proposal-id 123 --title "Pricing" --section-type cost
nusii sections update 456 --title "Updated Pricing"
nusii sections delete 456
```

### Line Items

```bash
nusii line-items list --section-id 456
nusii line-items get 789
nusii line-items create --section-id 456 --name "Design" --amount 50000 --quantity 1
nusii line-items update 789 --amount 60000
nusii line-items delete 789
```

Amounts are in **cents** (e.g., `50000` = $500.00).

### Other Commands

```bash
nusii activities list
nusii activities list --proposal-id 123
nusii users list
nusii themes list
nusii webhooks list
nusii webhooks get 123
nusii webhooks create --target-url "https://example.com/webhook" --events "proposal.sent,proposal.accepted"
nusii webhooks delete 123
```

## Output Formats

By default, output is a table when run interactively and JSON when piped.

```bash
# Force JSON output
nusii clients list -o json

# Force table output
nusii clients list -o table

# Pipe JSON to jq
nusii clients list -o json | jq '.data[].attributes.name'
```

## Configuration

Config file: `~/.config/nusii/config.yaml`

```yaml
api_key: "your-api-key"
api_url: "https://app.nusii.com"
output: "table"
```

Environment variables (`NUSII_API_KEY`, `NUSII_API_URL`, `NUSII_OUTPUT`) override the config file. Flags override everything.

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--api-key` | `-k` | API key |
| `--api-url` | | API base URL |
| `--output` | `-o` | Output format: `json` or `table` |
| `--no-input` | | Disable interactive prompts |
| `--debug` | | Print HTTP request/response details |

## Scripting & AI Agents

The CLI is designed to be scriptable and AI-agent friendly:

- JSON output when piped (`stdout` is not a TTY)
- `--no-input` disables all confirmation prompts (use `--confirm` on delete commands)
- Structured error output: `{"error": "message", "status": 404}`
- Consistent exit codes: `0` success, `1` error, `2` auth, `3` not found, `4` validation, `5` rate limited

## License

[Apache License 2.0](LICENSE)

## Development

```bash
make build      # Build binary to bin/nusii
make test       # Run unit tests
make lint       # Run go vet
make install    # Install to $GOPATH/bin
```
