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

- Run `go test -race -count=1 ./...` (or `task test`) before committing — CI uses `-count=1` to bypass the test cache
- Tests must pass with `-race` flag
- Use `testify` for assertions (`require` for fatal checks, `assert` for non-fatal)
- Use `iostreams.Test()` for testing command output
- Use `httptest.NewServer` for HTTP mocking
- E2E suites (`tests/e2e-*.sh`) hit a LIVE tenant via `./tests/run-all.sh` — need `zr auth login` first; some checks skip on the sandbox (see `docs/e2e-test-skips.md`)

## Architecture

- Follow the gh CLI patterns (see `docs/plans/README.md`)
- Commands go in `pkg/cmd/<resource>/<action>/`
- Internal packages go in `internal/`
- Use Factory for dependency injection
- Keep `cmd/zr/main.go` minimal

## Build & Run

- `task build` or `make build` — Build binary (output: `./bin/zr`)
- `task test` or `make test` — Run tests (`go test -race -count=1 ./...`)
- `task lint` or `make lint` — Run linters (go vet + staticcheck)
- `task fmt` or `make fmt` — `gofmt -w .` (run before pushing)
- `task check` or `make check` — lint + test (pre-commit gate)
- `task e2e` or `make e2e` — run E2E suites against a LIVE authenticated tenant
- Requires Go 1.26.1 (see `go.mod`)
- CI (`.github/workflows/ci.yml`) additionally enforces `gofmt -l .` and `go mod verify`, which `task lint`/`task check` do not — so run `task fmt` before pushing or CI fails on formatting
