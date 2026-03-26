package alias

import (
	"fmt"

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
			return runSet(f, args[0], args[1])
		},
	}
	return cmd
}

func runSet(f *factory.Factory, name, command string) error {
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
