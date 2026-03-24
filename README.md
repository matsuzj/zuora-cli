# zuora-cli (zr)

Zuora CLI — Work with Zuora from the command line.

## Installation

### From source

```bash
go install github.com/matsuzj/zuora-cli/cmd/zr@latest
```

## Quick Start

```bash
# Check version
zr version

# Generate shell completions
source <(zr completion bash)
```

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
│   │   └── completion/         # completion command
│   └── iostreams/              # I/O abstraction
├── internal/
│   └── build/                  # Build-time metadata
├── Taskfile.yml
├── Makefile
└── docs/plans/                 # Development plans
```

## License

MIT
