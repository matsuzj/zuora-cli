// Package globalflags is the single home of the CLI's persistent flags: their
// registration AND the PersistentPreRunE behavior that applies them (--env
// override, --read-only fail-safe, --zuora-version/--verbose client wiring,
// --json+--template rejection). pkg/cmd/root delegates here, and pkg/cmdtest
// applies the SAME logic to its stub root, so command tests exercise the real
// global-flag semantics instead of name-only stubs. (Extracted from root.go
// after a review caught that the test harness silently skipped this behavior.)
package globalflags

import (
	"fmt"
	"os"
	"strings"

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
	cmd.PersistentFlags().Bool("verbose", false, "Enable verbose/debug output")
	cmd.PersistentFlags().Bool("read-only", false, "Block write operations (POST/PUT/DELETE/PATCH)")
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
	if envName, _ := cmd.Flags().GetString("env"); envName != "" {
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
	verbose, _ := cmd.Flags().GetBool("verbose")

	// --read-only flag takes precedence over the ZR_READ_ONLY env var.
	readOnly, _ := cmd.Flags().GetBool("read-only")
	if !cmd.Flags().Changed("read-only") {
		readOnly = EnvReadOnly()
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
		if readOnly {
			client.SetReadOnly(true)
		}
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
