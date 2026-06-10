package main

import (
	"fmt"
	"strings"

	"github.com/google/shlex"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// aliasResolver is the subset of *alias.Store that alias expansion needs.
type aliasResolver interface {
	Get(name string) (string, bool)
}

// expandAlias returns args (an os.Args-shaped slice: args[0] is the binary
// name) with the first non-flag argument replaced by its alias expansion.
// For example, with "ls" aliased to "account list", ["zr","--json","ls"]
// becomes ["zr","--json","account","list"]. Names of commands registered on
// rootCmd are never expanded, and which global flags consume a value is
// derived from rootCmd's persistent flag definitions — both were previously
// hand-maintained lists that had drifted from root.go. When no expansion
// applies, args is returned unchanged. A malformed expansion (unbalanced
// quoting) returns the original args plus a non-nil error so the caller can
// warn and dispatch unexpanded.
func expandAlias(rootCmd *cobra.Command, args []string, store aliasResolver) ([]string, error) {
	if len(args) < 2 {
		return args, nil
	}

	builtins := builtinNames(rootCmd)
	takesValue := valueFlagSpellings(rootCmd)

	// Find the first non-flag argument (skip leading --flag and --flag=value)
	cmdIdx := -1
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			cmdIdx = i
			break
		}
		// Skip --flag value pairs (flags that take a value)
		if takesValue[arg] && i+1 < len(args) {
			i++ // skip the value
		}
	}
	if cmdIdx < 0 {
		return args, nil
	}

	cmdName := args[cmdIdx]

	// Don't expand built-in commands
	if builtins[cmdName] {
		return args, nil
	}

	expanded, ok := store.Get(cmdName)
	if !ok {
		return args, nil
	}

	// Split shell-style (like gh): quoted segments stay single arguments, so
	// an alias wrapping a ZOQL query survives intact. strings.Fields would
	// shred it.
	expandedArgs, err := shlex.Split(expanded)
	if err != nil {
		return args, fmt.Errorf("alias %q has a malformed expansion %q: %w", cmdName, expanded, err)
	}

	// Replace the alias at cmdIdx with expanded command words
	newArgs := make([]string, 0, len(args)+len(expandedArgs)-1)
	newArgs = append(newArgs, args[:cmdIdx]...)
	newArgs = append(newArgs, expandedArgs...)
	newArgs = append(newArgs, args[cmdIdx+1:]...)
	return newArgs, nil
}

// builtinNames returns every name dispatchable on rootCmd — registered command
// names and their cobra aliases, plus cobra's implicit "help" — so aliases can
// never shadow a real command. Derived, not hand-maintained: the old manual
// map had drifted (billrun/creditmemo/debitmemo were shadowable).
func builtinNames(rootCmd *cobra.Command) map[string]bool {
	names := map[string]bool{"help": true}
	for _, c := range rootCmd.Commands() {
		names[c.Name()] = true
		for _, a := range c.Aliases {
			names[a] = true
		}
	}
	return names
}

// valueFlagSpellings returns every spelling ("--name" and "-s") of rootCmd's
// persistent flags that consume a separate value argument (everything except
// bools). Derived from the live flag definitions instead of the old manual
// 5-spelling list copied from root.go.
func valueFlagSpellings(rootCmd *cobra.Command) map[string]bool {
	spellings := make(map[string]bool)
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if f.Value.Type() == "bool" {
			return
		}
		spellings["--"+f.Name] = true
		if f.Shorthand != "" {
			spellings["-"+f.Shorthand] = true
		}
	})
	return spellings
}
