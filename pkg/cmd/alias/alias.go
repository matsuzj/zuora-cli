package alias

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdAlias creates the "alias" command group.
func NewCmdAlias(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias <command>",
		Short: "Manage command aliases",
		Long:  "Create, delete, and list command aliases for zr.",
	}

	cmd.AddCommand(newCmdSet(f))
	cmd.AddCommand(newCmdDelete(f))
	cmd.AddCommand(newCmdList(f))

	return cmd
}
