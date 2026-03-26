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

// builtinCommands is the set of top-level command names that aliases must not shadow.
var builtinCommands = map[string]bool{
	"account": true, "alias": true, "api": true, "auth": true,
	"charge": true, "commitment": true, "completion": true, "config": true,
	"contact": true, "fulfillment": true, "fulfillment-item": true, "help": true,
	"invoice": true, "meter": true, "omnichannel": true, "order": true,
	"order-action": true, "order-line-item": true, "payment": true, "plan": true,
	"prepaid": true, "product": true, "query": true, "ramp": true,
	"rateplan": true, "signup": true, "subscription": true, "usage": true,
	"version": true,
}

// expandAliases loads aliases and rewrites os.Args if the first non-flag argument
// matches an alias name. For example, if "ls" is aliased to "account list",
// "zr --json ls" becomes "zr --json account list".
func expandAliases() {
	if len(os.Args) < 2 {
		return
	}

	// Find the first non-flag argument (skip leading --flag and --flag=value)
	cmdIdx := -1
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if !strings.HasPrefix(arg, "-") {
			cmdIdx = i
			break
		}
		// Skip --flag value pairs (flags that take a value)
		if (arg == "--env" || arg == "-e" || arg == "--zuora-version" ||
			arg == "--jq" || arg == "--template") && i+1 < len(os.Args) {
			i++ // skip the value
		}
	}
	if cmdIdx < 0 {
		return
	}

	cmdName := os.Args[cmdIdx]

	// Don't expand built-in commands
	if builtinCommands[cmdName] {
		return
	}

	store := alias.NewStore(config.Dir())
	if err := store.Load(); err != nil {
		return // silently ignore alias load failures
	}

	expanded, ok := store.Get(cmdName)
	if !ok {
		return
	}

	// Replace the alias at cmdIdx with expanded command words
	expandedArgs := strings.Fields(expanded)
	newArgs := make([]string, 0, len(os.Args)+len(expandedArgs)-1)
	newArgs = append(newArgs, os.Args[:cmdIdx]...)
	newArgs = append(newArgs, expandedArgs...)
	newArgs = append(newArgs, os.Args[cmdIdx+1:]...)
	os.Args = newArgs
}

func exitCode(err error) int {
	var ec exitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return 1
}
