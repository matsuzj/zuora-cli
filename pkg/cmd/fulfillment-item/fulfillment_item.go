// Package fulfillmentitem implements the "zr fulfillment-item" command group.
package fulfillmentitem

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/fulfillment-item/create"
	deletecmd "github.com/matsuzj/zuora-cli/pkg/cmd/fulfillment-item/delete"
	"github.com/matsuzj/zuora-cli/pkg/cmd/fulfillment-item/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/fulfillment-item/update"
	"github.com/spf13/cobra"
)

// NewCmdFulfillmentItem creates the fulfillment-item parent command.
func NewCmdFulfillmentItem(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fulfillment-item <command>",
		Short: "Manage Zuora fulfillment items",
		Long:  "Create, view, update, and delete Zuora fulfillment items.",
	}

	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))

	return cmd
}
