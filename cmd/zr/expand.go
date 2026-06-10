package main

import (
	"strings"
)

// aliasResolver is the subset of *alias.Store that alias expansion needs.
type aliasResolver interface {
	Get(name string) (string, bool)
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

// expandAlias returns args (an os.Args-shaped slice: args[0] is the binary
// name) with the first non-flag argument replaced by its alias expansion.
// For example, with "ls" aliased to "account list", ["zr","--json","ls"]
// becomes ["zr","--json","account","list"]. Built-in command names are never
// expanded; when no expansion applies, args is returned unchanged.
func expandAlias(args []string, store aliasResolver) []string {
	if len(args) < 2 {
		return args
	}

	// Find the first non-flag argument (skip leading --flag and --flag=value)
	cmdIdx := -1
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			cmdIdx = i
			break
		}
		// Skip --flag value pairs (flags that take a value)
		if (arg == "--env" || arg == "-e" || arg == "--zuora-version" ||
			arg == "--jq" || arg == "--template") && i+1 < len(args) {
			i++ // skip the value
		}
	}
	if cmdIdx < 0 {
		return args
	}

	cmdName := args[cmdIdx]

	// Don't expand built-in commands
	if builtinCommands[cmdName] {
		return args
	}

	expanded, ok := store.Get(cmdName)
	if !ok {
		return args
	}

	// Replace the alias at cmdIdx with expanded command words
	expandedArgs := strings.Fields(expanded)
	newArgs := make([]string, 0, len(args)+len(expandedArgs)-1)
	newArgs = append(newArgs, args[:cmdIdx]...)
	newArgs = append(newArgs, expandedArgs...)
	newArgs = append(newArgs, args[cmdIdx+1:]...)
	return newArgs
}
