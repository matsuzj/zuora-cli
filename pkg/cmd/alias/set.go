package alias

import (
	"fmt"

	"github.com/google/shlex"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func newCmdSet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <name> <command>",
		Short: "Create or update an alias",
		Long:  `Save a command alias. For example: zr alias set ls "account list"`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSet(cmd, f, args[0], args[1])
		},
	}
	return cmd
}

func runSet(cmd *cobra.Command, f *factory.Factory, name, command string) error {
	// Write-side guards. Dispatch never expands names that collide with real
	// commands, so accepting one here would silently create a dead alias (or,
	// before that dispatch guard covered every command, shadow the real one).
	// Validate at set time instead. `alias delete` deliberately stays
	// permissive so existing polluted entries can be removed.
	if isReservedName(cmd.Root(), name) {
		return fmt.Errorf("%q is a built-in command and cannot be aliased", name)
	}
	words, err := shlex.Split(command)
	if err != nil {
		return fmt.Errorf("malformed expansion %q: %w", command, err)
	}
	if len(words) == 0 {
		return fmt.Errorf("alias expansion must not be empty")
	}
	if words[0] == name {
		return fmt.Errorf("alias %q would invoke itself", name)
	}

	cfg, err := f.Config()
	if err != nil {
		return err
	}

	store := NewStore(cfg.ConfigDir())
	if err := store.Load(); err != nil {
		return err
	}

	store.Set(name, command)

	if err := store.Save(); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Alias %q set to %q\n", name, command)
	return nil
}

// isReservedName reports whether name is dispatchable on the root command
// (a registered command name, one of its cobra aliases, or the implicit
// help command) — the same set alias expansion refuses to expand.
func isReservedName(root *cobra.Command, name string) bool {
	if name == "help" {
		return true
	}
	for _, c := range root.Commands() {
		if c.Name() == name {
			return true
		}
		for _, a := range c.Aliases {
			if a == name {
				return true
			}
		}
	}
	return false
}
