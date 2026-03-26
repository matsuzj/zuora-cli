package alias

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func newCmdList(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all aliases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(f)
		},
	}
	return cmd
}

func runList(f *factory.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	store := NewStore(cfg.ConfigDir())
	if err := store.Load(); err != nil {
		return err
	}

	entries := store.All()
	if len(entries) == 0 {
		fmt.Fprintln(f.IOStreams.ErrOut, "No aliases configured.")
		return nil
	}

	out := f.IOStreams.Out
	for _, e := range entries {
		fmt.Fprintf(out, "%s\t%s\n", e.Name, e.Command)
	}
	return nil
}
