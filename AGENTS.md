# AGENTS.md - AI Agent Guidelines for zuora-cli

## Branch Naming

- `feature/<description>` — New features
- `fix/<description>` — Bug fixes
- `docs/<description>` — Documentation
- `chore/<description>` — Maintenance tasks

## Commit Messages

Follow Conventional Commits:

- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation
- `chore:` — Maintenance
- `test:` — Tests
- `refactor:` — Refactoring

## Branch Protection

- `main` branch is protected
- All changes must go through pull requests
- CI must pass before merge

## Go Code Standards

- Run `gofmt` on all files before committing
- Run `go vet ./...` and `staticcheck ./...` — fix all warnings
- All exported functions must have doc comments
- All new code must have tests (`*_test.go`)

## Testing

- Run `go test -race ./...` before committing
- Tests must pass with `-race` flag
- Use `testify/assert` for assertions
- Use `iostreams.Test()` for testing command output
- Use `httptest.NewServer` for HTTP mocking

## Architecture

- Follow the gh CLI patterns (see `docs/plans/README.md`)
- Commands go in `pkg/cmd/<resource>/<action>/`
- Internal packages go in `internal/`
- Use Factory for dependency injection
- Keep `cmd/zr/main.go` minimal

## Build & Run

- `task build` or `make build` — Build binary
- `task test` or `make test` — Run tests
- `task lint` or `make lint` — Run linters
- Binary output: `./bin/zr`
