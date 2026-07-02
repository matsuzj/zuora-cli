// Package globalflags is the single home of the CLI's persistent flags: their
// registration AND the PersistentPreRunE behavior that applies them (--env
// override, --read-only fail-safe, --zuora-version/--verbose client wiring,
// --json+--template rejection). pkg/cmd/root delegates here, and pkg/cmdtest
// applies the SAME logic to its stub root, so command tests exercise the real
// global-flag semantics instead of name-only stubs. (Extracted from root.go
// after a review caught that the test harness silently skipped this behavior.)
package globalflags

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// Register defines the global persistent flags on cmd — the one canonical
// list (alias expansion derives its flag-arity set from these definitions).
func Register(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("env", "e", "", "Environment name")
	cmd.PersistentFlags().Bool("json", false, "Output as JSON")
	cmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")
	cmd.PersistentFlags().String("template", "", "Format output with a Go template")
	cmd.PersistentFlags().Bool("csv", false, "Output as CSV")
	cmd.PersistentFlags().String("zuora-version", "", "Override Zuora API version header")
	cmd.PersistentFlags().CountP("verbose", "v", "Verbose output (-vv or ZR_DEBUG=api adds request/response bodies)")
	cmd.PersistentFlags().Bool("read-only", false, "Block write operations (POST/PUT/DELETE/PATCH)")
	cmd.PersistentFlags().Bool("read-only-allow-data-query", false, "In read-only mode, also allow Data Query submit/cancel (POST /query/jobs, DELETE /query/jobs/{id})")
	cmd.PersistentFlags().Duration("timeout", 0, "Abort if the command runs longer than this across all retries (e.g. 30s, 2m; 0 = no limit)")
}

// Apply is the PersistentPreRunE body shared by the real root and the test
// harness: it validates flag combinations and rewires the factory so the
// flags take effect for the command about to run.
func Apply(f *factory.Factory, cmd *cobra.Command) error {
	// --json + --template mutual exclusion. Other combinations are NOT
	// rejected: per the documented precedence (--jq implies JSON and wins;
	// --json/--jq/--template all win over --csv), the renderer deliberately
	// picks one, so e.g. `--json --jq .x` and `--csv --jq .x` are valid.
	jsonFlag, _ := cmd.Flags().GetBool("json")
	tmpl, _ := cmd.Flags().GetString("template")
	if jsonFlag && tmpl != "" {
		return fmt.Errorf("cannot use --json and --template together")
	}

	// default_output wiring (P4-3): when no output-format flag was given on
	// the command line AND the config says default_output=json AND stdout is
	// not a terminal, behave exactly as if --json had been passed. The TTY
	// guard keeps interactive sessions human-readable while scripts and
	// pipes get the configured machine default. A config load failure is
	// deliberately ignored here (the command itself surfaces it) — applying
	// a cosmetic default must never gate on config parsing (the alias
	// expansion lesson).
	// A subcommand could SHADOW a root persistent flag with a local one
	// (query's --csv did until P5-3c); cmd.Flags().Changed then consults the
	// local flag and misses
	// an explicit root-level `zr --csv query ...`, so check the root's
	// persistent flags too (review finding).
	formatFlagChanged := func(name string) bool {
		if cmd.Flags().Changed(name) {
			return true
		}
		if r := cmd.Root(); r != nil && r.PersistentFlags().Changed(name) {
			return true
		}
		return false
	}
	noFormatFlag := !formatFlagChanged("json") && !formatFlagChanged("jq") &&
		!formatFlagChanged("template") && !formatFlagChanged("csv")
	if noFormatFlag && f.Config != nil && f.IOStreams != nil && !f.IOStreams.IsTerminal() {
		if cfg, err := f.Config(); err == nil && cfg.DefaultOutput() == "json" {
			_ = cmd.Flags().Set("json", "true")
		}
	}

	// --env override (transient, does not persist to config.yml)
	// Environment selection: the --env flag wins over the ZR_ENV env var,
	// which wins over the config's active_environment (P6-4; same precedence
	// shape as --read-only vs ZR_READ_ONLY). ZR_ENV names an environment
	// EXACTLY — an unknown name errors just like --env, never a silent
	// fallback (a typo must not flip a write to another tenant).
	envName, _ := cmd.Flags().GetString("env")
	if envName == "" {
		envName = os.Getenv("ZR_ENV")
	}
	if envName != "" {
		origConfig := f.Config
		f.Config = func() (config.Config, error) {
			cfg, err := origConfig()
			if err != nil {
				return nil, err
			}
			// Validate environment exists
			if _, err := cfg.Environment(envName); err != nil {
				return nil, fmt.Errorf("invalid environment %q: %w", envName, err)
			}
			return &envOverrideConfig{Config: cfg, env: envName}, nil
		}
	}

	zv, _ := cmd.Flags().GetString("zuora-version")
	// Verbose levels (P6-3): -v = diagnostics, -vv (or ZR_DEBUG=api) also
	// logs request/response bodies. ZR_DEBUG=api implies level 1 so the
	// bodies appear in context.
	verboseCount, _ := cmd.Flags().GetCount("verbose")
	zrDebug := os.Getenv("ZR_DEBUG")
	verbose, verboseBody := VerboseLevels(verboseCount, zrDebug)
	// ZR_DEBUG is matched EXACTLY against "api" (VerboseLevels); any other
	// non-empty value — "API", "1", "true" — is a silent no-op that reads like
	// "debug is on" when it isn't. Warn so the user learns the only supported
	// spelling instead of wondering why bodies never appear. (#456)
	if dbg := strings.TrimSpace(zrDebug); dbg != "" && dbg != "api" && f.IOStreams != nil {
		fmt.Fprintf(f.IOStreams.ErrOut, "warning: ZR_DEBUG=%q is not recognized; the only supported value is \"api\" (adds request/response bodies at -vv). Ignoring.\n", dbg)
	}

	// --read-only flag takes precedence over the ZR_READ_ONLY env var.
	readOnly, _ := cmd.Flags().GetBool("read-only")
	if !cmd.Flags().Changed("read-only") {
		readOnly = EnvReadOnly()
	}

	// --read-only-allow-data-query flag takes precedence over its env var. This
	// opt-in only takes effect under read-only; off otherwise.
	allowDataQuery, _ := cmd.Flags().GetBool("read-only-allow-data-query")
	if !cmd.Flags().Changed("read-only-allow-data-query") {
		allowDataQuery = EnvReadOnlyAllowDataQuery()
	}

	// Apply all client overrides (context, version, verbose, read-only)
	// in a single wrapper captured from the original once, so the
	// overrides are not stacked cumulatively across invocations.
	// Auth observability (P6-2): the factory's lazy closures read
	// AuthLogWriter at call time, so setting it here covers both the
	// AuthToken path and the 401 force-refresh path.
	if verbose {
		f.AuthLogWriter = f.IOStreams.ErrOut
	}

	ctx := cmd.Context()
	// A positive --timeout bounds the WHOLE command (all retries + backoff), not
	// just one request. Apply runs as PersistentPreRunE, so it must NOT defer
	// cancel (that would fire before the command runs); instead it sets the
	// derived context on the command + client and releases the timer via the
	// root's PersistentPostRunE. No command sets one, so it always runs on the
	// success path; on the error path the process exits, so the timer cannot
	// outlive it. A deadline surfaces as context.DeadlineExceeded → exit 1,
	// distinct from Ctrl-C's context.Canceled → exit 130 (see cmd/zr/main.go).
	// Read the GLOBAL timeout from the ROOT's persistent flags, not cmd.Flags():
	// two subcommands define a LOCAL --timeout that shadows the inherited
	// persistent flag in the merged set — `order job-status` (its --watch poll
	// deadline) and `data-query run` (its submit+poll deadline) — so cmd.Flags()
	// would read the wrong one (Codex review finding). Each local flag keeps its
	// own meaning; the global whole-command deadline is set by
	// `zr --timeout … <cmd>`, which lands on the root persistent flag.
	var globalTimeout time.Duration
	if r := cmd.Root(); r != nil {
		globalTimeout, _ = r.PersistentFlags().GetDuration("timeout")
	}
	if globalTimeout > 0 {
		// cmd.Context() is non-nil under Execute/ExecuteContext, but guard the
		// nil parent (a command Apply'd directly without Execute) so WithTimeout
		// cannot panic — mirrors the `if ctx != nil` guard below.
		parent := ctx
		if parent == nil {
			parent = context.Background()
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(parent, globalTimeout)
		cmd.SetContext(ctx)
		root := cmd.Root()
		prev := root.PersistentPostRunE
		root.PersistentPostRunE = func(c *cobra.Command, args []string) error {
			cancel()
			if prev != nil {
				return prev(c, args)
			}
			return nil
		}
	}
	origHttpClient := f.HttpClient
	f.HttpClient = func() (*api.Client, error) {
		client, err := origHttpClient()
		if err != nil {
			return nil, err
		}
		if ctx != nil {
			client.SetContext(ctx)
		}
		if zv != "" {
			client.SetZuoraVersion(zv)
		}
		if verbose {
			client.SetVerbose(f.IOStreams.ErrOut)
		}
		if verboseBody {
			client.SetVerboseBody()
		}
		// Set both unconditionally so repeated Apply calls on a reused factory
		// are idempotent: a stale wrapper must never leave read-only state on.
		client.SetReadOnly(readOnly)
		client.SetReadOnlyAllowDataQuery(allowDataQuery)
		return client, nil
	}

	return nil
}

// EnvReadOnly reports whether ZR_READ_ONLY requests read-only mode. It accepts
// the conventional truthy/falsy spellings and, critically, fails safe: a
// non-empty but unrecognized value enables read-only rather than silently
// allowing writes.
func EnvReadOnly() bool {
	v := strings.TrimSpace(os.Getenv("ZR_READ_ONLY"))
	if v == "" {
		return false
	}
	switch strings.ToLower(v) {
	case "0", "f", "false", "no", "n", "off":
		return false
	default:
		// "1", "t", "true", "yes", "y", "on", and anything unrecognized.
		return true
	}
}

// EnvReadOnlyAllowDataQuery reports whether ZR_READ_ONLY_ALLOW_DATA_QUERY opts in
// to allowing Data Query writes (submit/cancel) in read-only mode. Unlike
// EnvReadOnly, it fails safe toward the RESTRICTIVE side: an unrecognized
// non-empty value parses as false (do NOT widen), because this knob relaxes the
// read-only guard and must never be enabled by an unclear value.
func EnvReadOnlyAllowDataQuery() bool {
	v := strings.TrimSpace(os.Getenv("ZR_READ_ONLY_ALLOW_DATA_QUERY"))
	if v == "" {
		return false
	}
	switch strings.ToLower(v) {
	case "1", "t", "true", "yes", "y", "on":
		return true
	default:
		// "0", "false", ..., and anything unrecognized → do not widen.
		return false
	}
}

// envOverrideConfig wraps a Config to override ActiveEnvironment() without mutating the original.
type envOverrideConfig struct {
	config.Config
	env string
}

func (c *envOverrideConfig) ActiveEnvironment() string { return c.env }

// SetActiveEnvironment delegates to the underlying config so that
// explicit "config set active_environment" still persists.
func (c *envOverrideConfig) SetActiveEnvironment(name string) error {
	return c.Config.SetActiveEnvironment(name)
}

// VerboseLevels derives the two verbose gates from the --verbose count and
// the ZR_DEBUG env value. ZR_DEBUG=api implies BOTH levels (bodies without
// surrounding diagnostics would lack context); any other value is ignored.
func VerboseLevels(count int, zrDebug string) (verbose, verboseBody bool) {
	debugAPI := zrDebug == "api"
	return count >= 1 || debugAPI, count >= 2 || debugAPI
}
