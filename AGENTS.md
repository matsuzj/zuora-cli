# AGENTS.md - AI Agent Guidelines for zuora-cli

Guidance for AI coding agents working in this repo. Read this first.

## Build & Run

- `task build` or `make build` — Build binary (output: `./bin/zr`, gitignored)
- `task test` or `make test` — Run tests (`go test -race -count=1 -coverprofile=cov.out -covermode=atomic ./...`)
- `task lint` or `make lint` — Linters (go vet + staticcheck + deadcode)
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
5. `make vuln` (i.e. `go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...`) — **CI runs govulncheck and fails on any vulnerability finding.** (Pinned to v1.3.0: v1.4.0's bundled x/tools v0.46.0 panics under the Go 1.26 toolchain. See the Makefile `vuln` comment.)
6. `go test -race -count=1 ./...`
7. Coverage floor: **≥ 83.0%** total (-coverpkg: covered-by-any-test semantics) (`make cover` enforces it locally; CI enforces it too)
8. `make build` (what CI runs — produces `bin/zr` with version ldflags; a bare `go build ./...` does not exercise that linkage)
9. For changes to live API/auth/output behavior: run the **E2E suite** (`make e2e` — every `tests/e2e-*.sh` suite against the live sandbox) — it is the only thing that catches Zuora-specific behavior that mocked unit tests miss. E2E is a MANUAL pre-merge/release gate and is intentionally NOT in CI.

`main` is protected (strict): PRs serialize. After one PR merges, others go BEHIND — `gh pr update-branch <n>`, wait for CI, then merge. `--admin`/auto-merge are not available.

## Go Code Standards

- `gofmt` all files; `go vet` + `go tool staticcheck` + `go tool deadcode` clean before committing.
- Exported functions need doc comments. New code needs tests (`*_test.go`).
- Use `testify` (`require` for fatal, `assert` for non-fatal).
- Command example invocations go in cobra's `Example:` field (or
  `listcmd.Spec.Example`), never embedded in `Long` — `make lint` rejects
  `Examples:` blocks inside `pkg/cmd` Go files.
- New commands are DECLARATIVE: `cmdutil.RunDetail`+`Action` (detail/write),
  `listcmd.New`+`Spec` (table lists), `output.RenderJSONOnly` (JSON-only).
  Hand-written `runE` only for the documented exceptions — see
  `docs/architecture.md`「コマンドの書き方(正準)」.
- **Every command honors the global format flags** (`--json`/`--jq`/`--template`/`--csv`),
  including local, non-API ones (`auth status`, `alias list`, `config list`).
  Never hand-roll output that ignores them — silently emitting text when
  `--json` was asked for is a bug (#453). Route through `pkg/output`; for
  non-API data, synthesize a JSON body and call `output.RenderDetail`/`Render`
  (see `version.go`).
- A write command that renders a JSON response body plus a stderr success line
  uses `output.RenderJSONWithMessage` — do not re-inline the
  `jq/template → PrintJSON → Fprintf` tail (it was deduped in #453). Bodyless
  writes use `output.RenderSuccess`.
- `output.Render` prints `No results found.` to stderr for a zero-row human
  table (stdout stays empty; `--json`/`--csv` are unaffected). Don't hand-roll
  empty-state checks in a command.
- Command options live in an options struct (`opts := &xxxOptions{Factory: f}`).
- Flag vocabulary: `--account-key` (ID or number, path param) /
  `--account-number` (`accountNumber` query) / `--account-id` (`accountId`
  query). Pick by what the ENDPOINT accepts; never reuse `--account`.
- The README destructive-command list is GENERATED: after adding/removing a
  `RequireConfirm` call, run `scripts/gen-destructive-list.sh` and refresh the
  block between the README's destructive-list markers (`make lint` fails on
  drift).

## Design decisions (settled — do not re-propose)

From the 2026-07 command-design audit. These were evaluated and deliberately kept; don't reopen them without a new reason:

- **Command layout stays `pkg/cmd/<resource>/<action>/`** (gh-CLI style). Do not move commands under a top-level `cmd/` tree — it is pure import churn with no behavior gain.
- **No Viper.** The custom `cobra`+`pflag`+`internal/config` mechanism already gives a single source of truth and documented `flag > env > config` precedence, and it does what `viper.BindPFlag` cannot: `ZR_READ_ONLY` fails **closed** while `ZR_READ_ONLY_ALLOW_DATA_QUERY` fails **restrictive** (asymmetric env parsing). Adopting Viper would regress that.
- **Output flags stay boolean** (`--json`/`--jq`/`--template`/`--csv`), not a single `--output` enum: it matches gh, and a global `--output` would collide with `data-query run`'s existing `--output` (result-file path).
- **Env vars stay `ZR_*`.** There is no `ZUORA_*` legacy to support, so renaming would only break existing users.

## Testing

- `iostreams.Test()` for command output; `httptest.NewServer` for HTTP mocking; `factory.NewTestFactory(ios, cfg, baseURL, token)` to wire a command to a mock server.
- **Build fixtures from REAL response shapes, not from memory.** A hand-written fixture that encodes the same key the command reads will pass even when both are wrong — this "fixture masking" has caused repeated silent-empty-field bugs (e.g. wrong/nested response keys). When fixing a response-mapping bug, **prove the test bites**: revert the production fix and confirm the test now FAILS, then restore.
- **A Detail-command `_Success` test must assert at least one rendered field value that is sourced from a NESTED response key** (e.g. `order.status`, `basicInfo.accountNumber`) — not merely `require.NoError` or the top-level `success:true`. Detail handlers commonly unwrap a nested object with a nil-fallback to the raw map (`order/get`, `ramp/get`, `fulfillment/get`, `account/get`); a fixture built with the wrong flat key then renders EMPTY in production yet still passes a test that only checks `NoError`. Asserting a nested value by name is what makes the bite-proof above actually bite. This rule is enforced in **review**, not by a lint: `cmdutil.RunDetail` backs both read-detail **and write** commands, and a write command's `_Success` test correctly asserts the **stderr** success message (and the request body) rather than a stdout field — so an automated "must assert a stdout field" check would false-positive on every write command, and heuristically separating the two is fragile. (See #308.)
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
- Branch naming: `feat/` `fix/` `docs/` `chore/` `refactor/` `test/` `perf/` `sec/`.
- Conventional Commits: `feat:` `fix:` `docs:` `chore:` `test:` `refactor:` `perf:` `ci:`.

## Reviewing a branch with sub-agents / tools

- Inspect a pushed branch READ-ONLY via `git diff main...origin/<branch>` and `git show origin/<branch>:<path>`. **Do NOT `git checkout` the branch in a shared worktree from a sub-agent/tool** — a stray `git checkout main` silently switches the whole tree and discards in-flight work. (Commit + push before any review tooling so a clobber is one `git checkout <branch>` away from recovery, and verify the branch/HEAD afterward — `git status` clean does not prove the tree is intact.)
- Second-opinion review with Codex: `codex exec review --base main` (the dedicated subcommand). Do NOT use freeform `codex exec "<prompt>"` (it hangs). Codex is non-deterministic — run a couple of passes for important changes.

## Architecture

- See `docs/architecture.md` for the current structure (gh CLI patterns). Commands in `pkg/cmd/<resource>/<action>/` use the declarative runners (`cmdutil.RunDetail` / `listcmd.Spec` — hand-written runE only for documented exceptions); infra in `internal/{api,auth,config}`; shared helpers in `pkg/cmdutil` and `pkg/output`. Use the Factory for DI; keep `cmd/zr/main.go` minimal.

## Release process

- Releases are cut by pushing a `vX.Y.Z` git tag → `.github/workflows/release.yml` runs GoReleaser (`release --clean`, pinned `~> v2`) → publishes a GitHub release (darwin/linux × amd64/arm64 + checksums) and updates the Homebrew tap.
- The tap ships a **formula** (`brews:`), NOT a cask — keeping macOS **and** Linux `brew install matsuzj/tap/zr` working. (A cask is macOS-only; do not migrate.)
- Releasing is an irreversible outward action — get explicit human sign-off first, and run the E2E gate on the exact release commit.
