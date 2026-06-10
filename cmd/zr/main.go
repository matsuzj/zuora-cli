package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/root"
)

func main() {
	f := factory.New()

	// Build the command tree first (construction is side-effect free), so
	// alias expansion can derive the builtin-name and flag-arity sets from it.
	rootCmd := root.NewCmdRoot(f)

	// Resolve aliases: expand os.Args before Cobra dispatch. The alias file
	// lives in the config dir; if the config itself cannot load, skip
	// expansion and let dispatch surface the real config error.
	if cfg, err := f.Config(); err == nil {
		os.Args = resolveAliasArgs(rootCmd, cfg.ConfigDir(), os.Args, f.IOStreams.ErrOut)
	}

	// Cancel in-flight requests and retry backoff on Ctrl-C / SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(f.IOStreams.ErrOut, "Error: %s\n", err)
		os.Exit(exitCode(err))
	}
}

type exitCoder interface {
	ExitCode() int
}

// expandAliases loads aliases and rewrites os.Args if the first non-flag argument
// matches an alias name. For example, if "ls" is aliased to "account list",
// "zr --json ls" becomes "zr --json account list".
func exitCode(err error) int {
	var ec exitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return 1
}
