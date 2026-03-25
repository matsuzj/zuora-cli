// Package orderaction implements the "zr order-action" command group.
package orderaction

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/order-action/update"
	"github.com/spf13/cobra"
)

// NewCmdOrderAction creates the order-action parent command.
func NewCmdOrderAction(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order-action <command>",
		Short: "Manage Zuora order actions",
		Long:  "Update and manage Zuora order actions.",
	}

	cmd.AddCommand(update.NewCmdUpdate(f))

	return cmd
}
