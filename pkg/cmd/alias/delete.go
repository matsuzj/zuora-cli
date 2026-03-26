package alias

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func newCmdDelete(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an alias",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(f, args[0])
		},
	}
	return cmd
}

func runDelete(f *factory.Factory, name string) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	store := NewStore(cfg.ConfigDir())
	if err := store.Load(); err != nil {
		return err
	}

	if err := store.Delete(name); err != nil {
		return err
	}

	if err := store.Save(); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Alias %q deleted\n", name)
	return nil
}
