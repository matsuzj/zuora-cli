# zuora-cli (zr)

Zuora CLI -- Work with Zuora from the command line.

A CLI tool for Zuora API operations, modeled after GitHub CLI (`gh`). Manage accounts, subscriptions, orders, invoices, payments, and more from your terminal.

## Installation

### Homebrew (macOS / Linux)

```bash
brew install matsuzj/tap/zr
```

### Binary releases

Download pre-built binaries from the [Releases](https://github.com/matsuzj/zuora-cli/releases) page.

### From source

```bash
go install github.com/matsuzj/zuora-cli/cmd/zr@latest
```

## Quick Start

```bash
# Authenticate
zr auth login

# List accounts
zr account list
zr account list --json
zr account list --jq '.data[].name'

# Get account details
zr account get A00000001

# List subscriptions for an account
zr subscription list --account A00000001

# Create an order
zr order create --body @order.json

# List invoices
zr invoice list --account A00000001

# Execute a ZOQL query
zr query "SELECT Id, Name, Status FROM Account WHERE Status = 'Active'"
zr query "SELECT Id FROM Invoice" --csv --export invoices.csv

# Raw API call
zr api /v1/accounts/A00000001
zr api /v1/orders -X POST --body @order.json
```

## Commands

| Command | Description |
|---------|-------------|
| `account` | Account CRUD + summary + payment methods |
| `subscription` | Subscription CRUD + lifecycle (cancel/suspend/resume/renew) + ChangeLog |
| `order` | Order CRUD + lifecycle (activate/cancel/revert) + async operations |
| `order-action` | Update order actions |
| `order-line-item` | Order line item CRUD + bulk update |
| `contact` | Contact CRUD + transfer + scrub + snapshot |
| `signup` | Create account + payment method + subscription in one call |
| `product` | Commerce Product CRUD |
| `plan` | Commerce Plan CRUD + purchase-options |
| `charge` | Commerce Charge CRUD + tiers update |
| `rateplan` | Get rate plan (v1 API) |
| `invoice` | Invoice list + get + items + files + email |
| `payment` | Payment list + get + create + apply + refund |
| `usage` | Usage record CRUD + CSV upload |
| `meter` | Meter run + debug + summary + audit trail |
| `ramp` | Ramp get + metrics |
| `commitment` | Commitment list + get + periods + balance + schedules |
| `fulfillment` | Fulfillment CRUD |
| `fulfillment-item` | Fulfillment item CRUD |
| `prepaid` | Prepaid balance operations (rollover/reverse/deplete) |
| `query` | ZOQL query execution (pagination + CSV/JSON export) |
| `omnichannel` | Omni-channel subscription create + get + delete |
| `alias` | Command alias management |
| `auth` | Authentication (login/logout/status/token) |
| `config` | Configuration management |
| `api` | Raw API requests |
| `version` | Print version |
| `completion` | Generate shell completion scripts |

## Global Flags

```
-e, --env <name>           Environment name (sandbox, us-production, etc. Run `zr config env list` to see all)
    --json                 Output as JSON
    --jq <expr>            Filter JSON output with a jq expression
    --template <tmpl>      Format output with a Go template
    --zuora-version <ver>  Override Zuora API version header
    --verbose              Enable debug output
```

**Output modes**: `--jq` (implies JSON), `--json`, and `--template` are mutually exclusive. Default output is a formatted table.

## Shell Completion

```bash
# Bash
source <(zr completion bash)

# Zsh
source <(zr completion zsh)

# Fish
zr completion fish | source
```

Add the appropriate line to your shell profile (`~/.bashrc`, `~/.zshrc`) to load completions on every session.

## Aliases

```bash
# Create aliases
zr alias set ls "account list"
zr alias set inv "invoice list --account"

# Use them
zr ls              # expands to: zr account list
zr --json ls       # expands to: zr --json account list (global flags supported)

# List and delete
zr alias list
zr alias delete ls
```

Aliases are stored in `~/.config/zr/aliases.yml`. Built-in command names cannot be overridden.

## Development

### Prerequisites

- Go 1.26+
- [go-task](https://taskfile.dev/) (`brew install go-task`)
- [staticcheck](https://staticcheck.dev/) (`go install honnef.co/go/tools/cmd/staticcheck@latest`)

### Build & Test

```bash
task build          # outputs ./bin/zr
task test           # go test -race ./...
task lint           # go vet + staticcheck
```

## License

MIT
