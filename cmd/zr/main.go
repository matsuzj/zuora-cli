package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/root"
	"github.com/matsuzj/zuora-cli/pkg/output"
)

func main() {
	f := factory.New()

	// Build the command tree first (construction is side-effect free), so
	// alias expansion can derive the builtin-name and flag-arity sets from it.
	rootCmd := root.NewCmdRoot(f)

	// Resolve aliases: expand os.Args before Cobra dispatch. Only the config
	// DIRECTORY is needed here (config.Dir() is pure path resolution, the same
	// XDG logic the loaded config uses) — deliberately NOT f.Config(): gating
	// expansion on a successful config parse would silently disable aliases
	// whenever any config file is malformed, even for aliases that target
	// commands needing no config at all.
	os.Args = resolveAliasArgs(rootCmd, config.Dir(), os.Args, f.IOStreams.ErrOut)

	// Cancel in-flight requests and retry backoff on Ctrl-C / SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(f.IOStreams.ErrOut, errorLine(err))
		os.Exit(exitCode(err))
	}
}

// errorLine renders err for stderr. API/OAuth error messages embed
// response-body text (Zuora reason messages, raw gateway pages), so the
// error path needs the same terminal-escape sanitization the stdout
// table/detail path already applies; newlines in multi-line errors survive.
func errorLine(err error) string {
	return "Error: " + output.SanitizeErrorText(err.Error())
}

type exitCoder interface {
	ExitCode() int
}

func exitCode(err error) int {
	// Ctrl-C: the conventional 128+SIGINT code, checked before exitCoder so
	// a wrapped cancellation is not misreported as an API failure.
	if errors.Is(err, context.Canceled) {
		return 130
	}
	var ec exitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return 1
}
