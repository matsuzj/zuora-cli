// Package payment implements the "zr payment" command group.
package payment

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/payment/apply"
	"github.com/matsuzj/zuora-cli/pkg/cmd/payment/cancel"
	"github.com/matsuzj/zuora-cli/pkg/cmd/payment/create"
	"github.com/matsuzj/zuora-cli/pkg/cmd/payment/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/payment/list"
	"github.com/matsuzj/zuora-cli/pkg/cmd/payment/refund"
	"github.com/matsuzj/zuora-cli/pkg/cmd/payment/transfer"
	"github.com/matsuzj/zuora-cli/pkg/cmd/payment/unapply"
	"github.com/spf13/cobra"
)

// NewCmdPayment creates the payment parent command.
func NewCmdPayment(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment <command>",
		Short: "Manage Zuora payments",
		Long:  "List, view, create, and manage Zuora payments.",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(apply.NewCmdApply(f))
	cmd.AddCommand(unapply.NewCmdUnapply(f))
	cmd.AddCommand(refund.NewCmdRefund(f))
	cmd.AddCommand(cancel.NewCmdCancel(f))
	cmd.AddCommand(transfer.NewCmdTransfer(f))

	return cmd
}
