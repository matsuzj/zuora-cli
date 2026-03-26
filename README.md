# zuora-cli (zr)

Zuora CLI -- Work with Zuora from the command line.

## Installation

### Homebrew (macOS / Linux)

```bash
brew install matsuzj/tap/zr
```

### From source

```bash
go install github.com/matsuzj/zuora-cli/cmd/zr@latest
```

### Binary releases

Download pre-built binaries from the [Releases](https://github.com/matsuzj/zuora-cli/releases) page.

## Quick Start

```bash
# Check version
zr version

# Authenticate (interactive)
zr auth login

# Set the active environment
zr config set active_environment sandbox

# List accounts
zr account list

# Get a subscription
zr subscription get SUB-00000001

# Raw API call
zr api get /v1/accounts
```

## Shell Completion

```bash
# Bash
source <(zr completion bash)

# Zsh
source <(zr completion zsh)

# Fish
zr completion fish | source
```

To load completions on every session, add the appropriate line to your shell profile (e.g. `~/.bashrc`, `~/.zshrc`).

## Aliases

Save frequently used commands as aliases:

```bash
# Create an alias
zr alias set ls "account list"

# List all aliases
zr alias list

# Delete an alias
zr alias delete ls
```

Aliases are stored in `~/.config/zr/aliases.yml`.

## Development

### Prerequisites

- Go 1.26+
- [go-task](https://taskfile.dev/) (optional, `brew install go-task`)
- [staticcheck](https://staticcheck.dev/) (`go install honnef.co/go/tools/cmd/staticcheck@latest`)

### Build

```bash
# With task
task build

# With make (fallback)
make build

# Binary is output to ./bin/zr
./bin/zr version
```

### Test

```bash
task test    # or: make test
```

### Lint

```bash
task lint    # or: make lint
```

## Project Structure

```
zuora-cli/
├── cmd/zr/main.go              # Entrypoint
├── pkg/
│   ├── cmd/
│   │   ├── root/               # Root command + global flags
│   │   ├── factory/            # DI (IOStreams, Config, HTTPClient)
│   │   ├── version/            # version command
│   │   ├── completion/         # completion command
│   │   └── alias/              # alias set/delete/list
│   └── iostreams/              # I/O abstraction
├── internal/
│   ├── build/                  # Build-time metadata
│   └── config/                 # Config file management
├── .goreleaser.yml             # Release configuration
├── Taskfile.yml
├── Makefile
└── docs/plans/                 # Development plans
```

## License

MIT
