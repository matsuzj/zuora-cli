// Package account implements the "zr account" command group.
package account

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/account/create"
	deletecmd "github.com/matsuzj/zuora-cli/pkg/cmd/account/delete"
	"github.com/matsuzj/zuora-cli/pkg/cmd/account/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/account/list"
	paymentmethods "github.com/matsuzj/zuora-cli/pkg/cmd/account/payment-methods"
	paymentmethodscascading "github.com/matsuzj/zuora-cli/pkg/cmd/account/payment-methods-cascading"
	paymentmethodsdefault "github.com/matsuzj/zuora-cli/pkg/cmd/account/payment-methods-default"
	setcascading "github.com/matsuzj/zuora-cli/pkg/cmd/account/set-cascading"
	"github.com/matsuzj/zuora-cli/pkg/cmd/account/summary"
	"github.com/matsuzj/zuora-cli/pkg/cmd/account/update"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdAccount creates the account parent command.
func NewCmdAccount(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account <command>",
		Short: "Manage Zuora accounts",
		Long:  "List, view, create, update, and delete Zuora billing accounts.",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(summary.NewCmdSummary(f))
	cmd.AddCommand(paymentmethods.NewCmdPaymentMethods(f))
	cmd.AddCommand(paymentmethodsdefault.NewCmdPaymentMethodsDefault(f))
	cmd.AddCommand(paymentmethodscascading.NewCmdPaymentMethodsCascading(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))
	cmd.AddCommand(setcascading.NewCmdSetCascading(f))

	return cmd
}
