package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/alias"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/root"
)

func main() {
	f := factory.New()

	// Resolve aliases: expand os.Args before Cobra dispatch
	expandAliases()

	rootCmd := root.NewCmdRoot(f)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(f.IOStreams.ErrOut, "Error: %s\n", err)
		os.Exit(exitCode(err))
	}
}

type exitCoder interface {
	ExitCode() int
}

// expandAliases loads aliases and rewrites os.Args if the first non-flag argument
// matches an alias name. For example, if "ls" is aliased to "account list",
// "zr ls --json" becomes "zr account list --json".
func expandAliases() {
	if len(os.Args) < 2 {
		return
	}
	// Don't expand if the first arg is a built-in command or starts with -
	firstArg := os.Args[1]
	if strings.HasPrefix(firstArg, "-") {
		return
	}

	store := alias.NewStore(config.Dir())
	if err := store.Load(); err != nil {
		return // silently ignore alias load failures
	}

	expanded, ok := store.Get(firstArg)
	if !ok {
		return
	}

	// Replace the alias with expanded command words
	expandedArgs := strings.Fields(expanded)
	newArgs := make([]string, 0, 1+len(expandedArgs)+len(os.Args)-2)
	newArgs = append(newArgs, os.Args[0])
	newArgs = append(newArgs, expandedArgs...)
	newArgs = append(newArgs, os.Args[2:]...)
	os.Args = newArgs
}

func exitCode(err error) int {
	var ec exitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return 1
}
