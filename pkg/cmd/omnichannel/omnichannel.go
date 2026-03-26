// Package omnichannel implements the "zr omnichannel" command group.
package omnichannel

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/omnichannel/create"
	deletecmd "github.com/matsuzj/zuora-cli/pkg/cmd/omnichannel/delete"
	"github.com/matsuzj/zuora-cli/pkg/cmd/omnichannel/get"
	"github.com/spf13/cobra"
)

// NewCmdOmnichannel creates the omnichannel parent command.
func NewCmdOmnichannel(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "omnichannel <command>",
		Short: "Manage Zuora omni-channel subscriptions",
		Long:  "Create, view, and delete Zuora omni-channel subscriptions.",
	}

	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))

	return cmd
}
