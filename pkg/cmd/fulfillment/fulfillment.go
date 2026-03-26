// Package fulfillment implements the "zr fulfillment" command group.
package fulfillment

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/fulfillment/create"
	deletecmd "github.com/matsuzj/zuora-cli/pkg/cmd/fulfillment/delete"
	"github.com/matsuzj/zuora-cli/pkg/cmd/fulfillment/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/fulfillment/update"
	"github.com/spf13/cobra"
)

// NewCmdFulfillment creates the fulfillment parent command.
func NewCmdFulfillment(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fulfillment <command>",
		Short: "Manage Zuora fulfillments",
		Long:  "Create, view, update, and delete Zuora fulfillments.",
	}

	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))

	return cmd
}
