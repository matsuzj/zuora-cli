# zuora-cli (zr)

Zuora CLI -- Work with Zuora from the command line.

A CLI tool for Zuora API operations, modeled after GitHub CLI (`gh`). Manage accounts, subscriptions, orders, invoices, payments, and more from your terminal.

## Installation

### Homebrew (macOS / Linux)

```bash
brew install matsuzj/tap/zr
```

Shell completions (bash, zsh, fish) are installed automatically.

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
| `account` | Account CRUD + summary + payment-methods (default/cascading) + set-cascading |
| `subscription` | Subscription CRUD + lifecycle (cancel/suspend/resume/renew) + ChangeLog |
| `order` | Order CRUD + lifecycle (activate/cancel/revert) + async operations |
| `order-action` | Update order actions |
| `order-line-item` | Order line item get/update + bulk update |
| `contact` | Contact CRUD + transfer + scrub + snapshot |
| `signup` | Create account + payment method + subscription in one call |
| `product` | Commerce Product create/get/update + list-legacy |
| `plan` | Commerce Plan create/get/list/update + purchase-options |
| `charge` | Commerce Charge CRUD + tiers update |
| `rateplan` | Get rate plan (v1 API) |
| `invoice` | Invoice list + get + items + files + email + usage-rate-detail |
| `creditmemo` | Credit memo list + get (filter by account/status) |
| `debitmemo` | Debit memo list + get (filter by account/status) |
| `billrun` | Bill run create + get + post + cancel + delete |
| `payment` | Payment list + get + create + apply + refund |
| `usage` | Usage record CRUD + CSV upload |
| `meter` | Meter run + debug + status + summary + audit |
| `ramp` | Ramp get/get-by-subscription + metrics/metrics-by-order/metrics-by-subscription |
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
-e, --env <name>           Environment name (sandbox, us-production, eu-production, apac-production, etc.)
    --json                 Output as JSON
    --jq <expr>            Filter JSON output with a jq expression
    --template <tmpl>      Format output with a Go template
    --csv                  Output as CSV (list/table commands)
    --zuora-version <ver>  Override Zuora API version header
    --verbose              Enable debug output
    --read-only            Block write operations (POST/PUT/DELETE/PATCH)
```

**Output modes**: `--json` and `--template` are mutually exclusive. `--jq` implies JSON output and takes precedence when combined with other flags. `--csv` renders list/table output as CSV (and detail output as a `Field,Value` table) with spreadsheet formula-injection sanitization; `--json` / `--jq` / `--template` take precedence over it. Default output is a formatted table.

**Read-only mode**: `--read-only` (or `ZR_READ_ONLY`) blocks all write operations (PUT/DELETE/PATCH and most POST requests). The environment variable accepts any conventional truthy value (`true`, `1`, `yes`, `on`); for safety it **fails closed** — a non-empty value that isn't a recognized falsy spelling (`false`, `0`, `no`, `off`) enables read-only rather than silently allowing writes. The `--read-only` flag takes precedence over the env var. Read-only POST endpoints — ZOQL queries, Commerce API queries/lists, order/subscription previews, and meter summaries — are allowed. See [docs/plans/read-only-mode.md](docs/plans/read-only-mode.md) for the full allowlist.

**Destructive operations**: irreversible commands require an explicit `--confirm` flag. This includes `delete` for `account` / `contact` / `order` / `subscription` / `usage` / `fulfillment` / `fulfillment-item` / `omnichannel` / `billrun`, plus `order cancel`, `order revert`, `subscription cancel`, `billrun cancel`, `contact scrub`, and `prepaid reverse-rollover`.

**Interrupts**: pressing Ctrl-C (SIGINT/SIGTERM) cancels any in-flight request and aborts retry backoff. Mutating requests (POST/PATCH) carry an `Idempotency-Key` header so a network retry cannot create a duplicate order, payment, or refund.

## Authentication

`zr` authenticates with Zuora using the OAuth 2.0 **client credentials** grant.
Create an OAuth client in your Zuora tenant to obtain a **Client ID** and **Client
Secret**, then provide them in one of three ways — resolved in this order:

1. `--client-id` / `--client-secret` flags
2. `ZR_CLIENT_ID` / `ZR_CLIENT_SECRET` environment variables
3. Interactive prompts (only when stdin is a terminal)

Authentication always targets the **active environment** (see
[Configuration](#configuration)); override it for a single command with
`-e/--env <name>`.

### Interactive login

```bash
zr auth login                    # prompts for Client ID and Client Secret
zr auth login -e us-production   # log in to a specific environment
```

Credentials are validated by fetching a token first (invalid credentials are
never stored), then saved to your OS keyring, scoped per environment. The Client
Secret prompt is masked.

### Login with flags

```bash
zr auth login --client-id <id> --client-secret <secret>
zr auth login --client-id <id> --client-secret <secret> -e us-production
```

Handy for scripted setup. As with interactive login, valid credentials are
persisted to the keyring.

### Login with environment variables (headless / CI)

On systems without an OS keyring (CI runners, containers, headless servers), set
the credentials as environment variables. When **both** are set they take
precedence over the keyring, and you do **not** need to run `zr auth login` at all
— any command obtains and caches an access token on demand:

```bash
export ZR_CLIENT_ID=your_client_id
export ZR_CLIENT_SECRET=your_client_secret

zr account list                  # authenticates automatically
zr -e us-production account list # same credentials, different environment
zr auth token                    # print the access token for use in other tools
```

Notes:

- **Both** variables must be set — a single one is ignored and falls back to the
  keyring / prompts.
- Env-var credentials are **not** environment-specific: the same Client ID/Secret
  are used for whichever environment you target via `-e/--env` or
  `active_environment`. (Keyring credentials, by contrast, are stored per
  environment.) Make sure the credentials are valid for the environment you select.
- You can still run `zr auth login` with the env vars set — it skips the prompts
  and additionally tries to copy the credentials into the keyring.

### Other auth commands

```bash
zr auth status   # active env, base URL, credential source (keyring vs env vars), token validity
zr auth token    # print the current access token (refreshes if expired) — for scripts
zr auth logout   # remove keyring credentials and the cached token for the active env
```

`zr auth logout` does **not** unset `ZR_CLIENT_ID`/`ZR_CLIENT_SECRET`; unset those
yourself to fully de-authenticate.

## Configuration

`zr` stores its configuration as YAML files under a per-user config directory. By
default this is `~/.config/zr/`; set `XDG_CONFIG_HOME` to relocate it (the files
then live under `$XDG_CONFIG_HOME/zr/`). The directory and its files are created
automatically on first write (e.g. `zr auth login` or `zr config set`) with
`0700`/`0600` permissions — you do not need to create them by hand.

### Config directory layout

```
~/.config/zr/                 # or $XDG_CONFIG_HOME/zr/
├── config.yml                # active environment, API version, default output
├── environments.yml          # custom / overridden environment definitions
├── tokens.yml                # cached OAuth access tokens (managed by `zr auth`)
└── aliases.yml               # command aliases (see Aliases below)
```

Missing files are fine — built-in defaults apply. A file that exists but is
malformed is reported as an error rather than silently ignored.

### `config.yml`

```yaml
active_environment: sandbox   # default: sandbox
zuora_version: "2025-08-12"   # default API version header (YYYY-MM-DD)
default_output: table         # table | json (default: table)
```

Manage these values with the `config` command (which writes the file for you):

```bash
zr config list                          # show all current values
zr config get active_environment        # read a single value
zr config set default_output json       # write a value
zr config set zuora_version 2025-08-12
zr config env us-production             # switch the active environment
```

### `environments.yml`

Built-in environments (`sandbox`, `apac-sandbox`, `us-production`,
`us-production-cloud2`, `eu-production`, `apac-production`) are available without
any configuration. To add or override one, edit `environments.yml`:

```yaml
environments:
  my-tenant:
    base_url: https://rest.na.zuora.com   # absolute http(s) URL, required
```

Select an environment per-invocation with `-e/--env <name>`, or persistently with
`zr config set active_environment <name>` / `zr config env <name>`.

### Credentials

Client credentials are **not** stored in the config directory. By default they
live in the OS keyring (`zr auth login`), or are read from `ZR_CLIENT_ID` /
`ZR_CLIENT_SECRET`. See [Authentication](#authentication) for details. Cached
OAuth access tokens (`tokens.yml`) are written automatically and are safe to
delete — they will be re-fetched on the next request.

### Precedence

For each setting, values are resolved highest-to-lowest:

**command-line flag** > **environment variable** > **config file** > **built-in default**

| Setting | Flag | Env var | Config file key |
|---------|------|---------|-----------------|
| Environment | `-e, --env` | — | `config.yml: active_environment` |
| API version | `--zuora-version` | — | `config.yml: zuora_version` |
| Output format | `--json` / `--template` | — | `config.yml: default_output` |
| Read-only | `--read-only` | `ZR_READ_ONLY` | — |
| Credentials | `--client-id` / `--client-secret` | `ZR_CLIENT_ID` / `ZR_CLIENT_SECRET` | OS keyring |
| Config dir | — | `XDG_CONFIG_HOME` | — |

### Environment variables

| Variable | Purpose |
|----------|---------|
| `ZR_CLIENT_ID` | OAuth client ID (with `ZR_CLIENT_SECRET`, overrides the keyring) |
| `ZR_CLIENT_SECRET` | OAuth client secret |
| `ZR_READ_ONLY` | Block write operations — truthy values enable it; fails closed (see [Read-only mode](#global-flags) and [docs/plans/read-only-mode.md](docs/plans/read-only-mode.md)) |
| `XDG_CONFIG_HOME` | Relocate the config directory (defaults to `~/.config`) |
| `NO_COLOR` | Disable colored output when set (any value) |
| `PAGER` | Pager command for long output (default: `less`); `LESS` / `LV` tune the respective pagers |

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

Aliases are stored in `$XDG_CONFIG_HOME/zr/aliases.yml` (defaults to `~/.config/zr/aliases.yml`). Built-in command names cannot be overridden.

## Development

### Prerequisites

- Go 1.26+
- [go-task](https://taskfile.dev/) (`brew install go-task`)
- [staticcheck](https://staticcheck.dev/) (`go install honnef.co/go/tools/cmd/staticcheck@latest`)

### Build & Test

```bash
task build          # outputs ./bin/zr
task test           # go test -race -count=1 ./...
task lint           # go vet + staticcheck
task fmt            # gofmt -w .
task check          # lint + test (pre-commit gate)
```

CI additionally enforces a `gofmt -l .` formatting gate and `go mod verify`, which
`task lint`/`task check` do not run — so run `task fmt` (or `gofmt -w .`) before
pushing, or CI will fail on formatting even when local checks pass.

End-to-end suites (run the real binary against a live Zuora tenant) live in
`tests/e2e-*.sh`; run them with `./tests/run-all.sh` after `zr auth login`. Some
checks legitimately skip on the sandbox tenant — see
[docs/e2e-test-skips.md](docs/e2e-test-skips.md) for each skip and its cause.

## License

MIT
