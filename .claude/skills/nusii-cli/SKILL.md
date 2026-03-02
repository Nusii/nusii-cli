---
name: nusii-cli
description: Use when the user wants to manage Nusii proposals, clients, sections, line items, or other Nusii resources. Translates natural language requests into nusii CLI commands.
argument-hint: "<natural language request>"
allowed-tools: Bash
---

# Nusii CLI Skill

You help users interact with the Nusii proposal software API using the `nusii` CLI tool.

## Available Commands

```
nusii account                          # Show account info
nusii auth login/status/logout         # Manage authentication
nusii clients list/get/create/update/delete
nusii proposals list/get/create/update/delete/send/archive
nusii sections list/get/create/update/delete
nusii line-items list/get/create/update/delete
nusii activities list/get
nusii users list
nusii themes list
nusii webhooks list/get/create/delete
```

## Global Flags

- `-o json` or `-o table` — output format (auto-detects TTY)
- `-k <key>` / `--api-key` — API key override
- `--api-url <url>` — API URL override (e.g., `http://localhost:3000`)
- `--no-input` — disable interactive prompts
- `--debug` — print HTTP details to stderr

## Key Flags by Command

**clients create/update:** `--name`, `--email`, `--surname`, `--business`, `--currency`, `--locale`, `--web`, `--telephone`, `--address`, `--city`, `--postcode`, `--country`, `--state`

**proposals list:** `--status <draft|pending|accepted|rejected>`, `--archived`, `--page`, `--per-page`
**proposals create:** `--title`, `--client-id`, `--client-email`, `--template-id`, `--theme`, `--currency`, `--expires-at`, `--display-date`, `--report`, `--exclude-total`
**proposals send:** `--email`, `--cc`, `--bcc`, `--subject`, `--message`

**sections list:** `--proposal-id`, `--template-id`, `--include-line-items`, `--page`, `--per-page`
**sections create/update:** `--proposal-id`, `--template-id`, `--title`, `--name`, `--body`, `--position`, `--section-type`, `--reusable`, `--optional`, `--include-total`, `--page-break`

**line-items create:** `--section-id` (required), `--name`, `--cost-type`, `--recurring-type`, `--per-type`, `--quantity`, `--amount` (in cents)

**webhooks create:** `--target-url`, `--events` (comma-separated)

**delete commands:** `--confirm` to skip confirmation prompt

## Examples

```bash
# Authenticate (saves key to ~/.config/nusii/config.yaml)
nusii auth login --api-key YOUR_API_KEY
```

## Behavior

- Always use `-o json` when you need to parse output or chain commands.
- Use `--confirm` on delete commands to skip prompts.
- When creating a proposal for a client by name, first look up the client ID with `nusii clients list -o json`.
- Amounts for line items are in **cents** (e.g., $100.00 = 10000).
- Show the user the command you're running before executing it.

## Request

$ARGUMENTS
