// Package orderlineitem implements the "zr order-line-item" command group.
package orderlineitem

import (
	bulkupdate "github.com/matsuzj/zuora-cli/pkg/cmd/order-line-item/bulk-update"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/order-line-item/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/order-line-item/update"
	"github.com/spf13/cobra"
)

// NewCmdOrderLineItem creates the order-line-item parent command.
func NewCmdOrderLineItem(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order-line-item <command>",
		Short: "Manage Zuora order line items",
		Long:  "View, update, and bulk-update Zuora order line items.",
	}

	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(bulkupdate.NewCmdBulkUpdate(f))

	return cmd
}
