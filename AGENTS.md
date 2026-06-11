# AGENTS.md - AI Agent Guidelines for zuora-cli

Guidance for AI coding agents working in this repo. Read this first.

## Build & Run

- `task build` or `make build` — Build binary (output: `./bin/zr`, gitignored)
- `task test` or `make test` — Run tests (`go test -race -count=1 ./...`)
- `task lint` or `make lint` — Linters (go vet + staticcheck)
- `task fmt` or `make fmt` — `gofmt -w .` (run before pushing)
- `task check` / `make check` — local pre-commit gate (see "Verifying changes" — it is a SUBSET of CI)
- `task e2e` or `make e2e` — run E2E suites against a LIVE authenticated tenant (`./tests/run-all.sh`)
- Requires the Go toolchain pinned in `go.mod` (**go 1.26.4**). If your shell's `go` resolves to a different/older version, invoke the matching toolchain explicitly via your version manager (e.g. an explicit `GOROOT=.../1.26.4` + `PATH`); a stale `go` will fail the build/tests in confusing ways.

## Verifying changes (match CI before you push)

CI (`.github/workflows/ci.yml`) gates merges on more than `make check` does. To avoid a red PR, run the **same** checks locally before pushing:

1. `go mod verify`
2. `gofmt -l .` — must print nothing (CI fails on any unformatted file). Run `make fmt` to auto-fix before pushing. (`make check`/`make ci` run this gate via `fmtcheck`; bare `make lint` does not.)
3. `go vet ./...`
4. `go tool staticcheck ./...` — **CI runs this; fix every finding.** The version is pinned by go.mod's `tool` directive, so local and CI always run the same staticcheck — no separate install. (Note: a `map[string]interface{}` → `any` editor hint is gopls "modernize", NOT staticcheck, and does not fail CI — the codebase uses `interface{}` throughout.)
5. `make vuln` (i.e. `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`) — **CI runs govulncheck and fails on any vulnerability finding.**
6. `go test -race -count=1 ./...`
7. Coverage floor: **≥ 73.0%** total (`make cover` enforces it locally; CI enforces it too)
8. `make build` (what CI runs — produces `bin/zr` with version ldflags; a bare `go build ./...` does not exercise that linkage)
9. For changes to live API/auth/output behavior: run the **E2E suite** (`make e2e`, 9 suites against the live sandbox) — it is the only thing that catches Zuora-specific behavior that mocked unit tests miss. E2E is a MANUAL pre-merge/release gate and is intentionally NOT in CI.

`main` is protected (strict): PRs serialize. After one PR merges, others go BEHIND — `gh pr update-branch <n>`, wait for CI, then merge. `--admin`/auto-merge are not available.

## Go Code Standards

- `gofmt` all files; `go vet` + `go tool staticcheck` clean before committing.
- Exported functions need doc comments. New code needs tests (`*_test.go`).
- Use `testify` (`require` for fatal, `assert` for non-fatal).

## Testing

- `iostreams.Test()` for command output; `httptest.NewServer` for HTTP mocking; `factory.NewTestFactory(ios, cfg, baseURL, token)` to wire a command to a mock server.
- **Build fixtures from REAL response shapes, not from memory.** A hand-written fixture that encodes the same key the command reads will pass even when both are wrong — this "fixture masking" has caused repeated silent-empty-field bugs (e.g. wrong/nested response keys). When fixing a response-mapping bug, **prove the test bites**: revert the production fix and confirm the test now FAILS, then restore.
- Zuora responses often NEST (e.g. under `basicInfo`/`ramp`/`fulfillment`) or return BULK arrays (`fulfillments[]`); do not assume flat top-level keys. Verify the real shape against the Zuora API reference and/or a live probe (`zr api <path>`).
- E2E: some checks legitimately skip on the sandbox (unprovisioned features); see `docs/e2e-test-skips.md`.

## Response handling & safety conventions

When adding/maintaining a command that calls the API:

- The Zuora success-flag check is **ON BY DEFAULT** in the API client: HTTP 200 with `{"success":false}` (or Object-CRUD `{"Success":false}`) becomes a non-zero exit, and it is a no-op for bodies without the flag. Do NOT pass `api.WithoutCheckSuccess()` in typed commands — it exists solely for the raw `zr api` GET/HEAD passthrough, which must deliver bodies uninterpreted. (This used to be an opt-in, `WithCheckSuccess()`, and missing call sites were a recurring bug class — the default flip made that structurally impossible.)
- Destructive/irreversible commands must gate on `cmdutil.RequireConfirm(confirm)` behind a `--confirm` flag (returns the canonical "this action is irreversible…" error). Do not inline the guard string — call the helper.
- Render response fields via `cmdutil.GetString` (plain) / `cmdutil.GetMoney` (monetary — fixed two decimals, the display contract) / `cmdutil.GetDecimal` (non-monetary numerics, avoids scientific notation) / `cmdutil.GetBool`/`GetInt`, descending into nested objects/array elements as the real response requires. Register `--body`/`--confirm` via `cmdutil.AddBodyFlag`/`AddConfirmFlag`, never hand-rolled.
- Zuora rejects **PUT requests carrying an `Idempotency-Key`** — the client adds the key to POST/PATCH only; do not change that.

## Git hygiene

- **Never `git add -A` / `git add .`** — the harness can drop stray files (e.g. `.claude/scheduled_tasks.lock`) into the worktree, and they get committed. Stage **explicit paths**.
- Branch naming: `feature/` `fix/` `docs/` `chore/` `refactor/` `test/`.
- Conventional Commits: `feat:` `fix:` `docs:` `chore:` `test:` `refactor:`.

## Reviewing a branch with sub-agents / tools

- Inspect a pushed branch READ-ONLY via `git diff main...origin/<branch>` and `git show origin/<branch>:<path>`. **Do NOT `git checkout` the branch in a shared worktree from a sub-agent/tool** — a stray `git checkout main` silently switches the whole tree and discards in-flight work. (Commit + push before any review tooling so a clobber is one `git checkout <branch>` away from recovery, and verify the branch/HEAD afterward — `git status` clean does not prove the tree is intact.)
- Second-opinion review with Codex: `codex exec review --base main` (the dedicated subcommand). Do NOT use freeform `codex exec "<prompt>"` (it hangs). Codex is non-deterministic — run a couple of passes for important changes.

## Architecture

- Follow gh CLI patterns (see `docs/plans/README.md`). Commands in `pkg/cmd/<resource>/<action>/`; infra in `internal/{api,auth,config}`; shared helpers in `pkg/cmdutil` and `pkg/output`. Use the Factory for DI; keep `cmd/zr/main.go` minimal.

## Release process

- Releases are cut by pushing a `vX.Y.Z` git tag → `.github/workflows/release.yml` runs GoReleaser (`release --clean`, pinned `~> v2`) → publishes a GitHub release (darwin/linux × amd64/arm64 + checksums) and updates the Homebrew tap.
- The tap ships a **formula** (`brews:`), NOT a cask — keeping macOS **and** Linux `brew install matsuzj/tap/zr` working. (A cask is macOS-only; do not migrate.)
- Releasing is an irreversible outward action — get explicit human sign-off first, and run the E2E gate on the exact release commit.
