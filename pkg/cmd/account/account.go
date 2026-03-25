// Package account implements the "zr account" command group.
package account

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/account/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/account/list"
	paymentmethods "github.com/matsuzj/zuora-cli/pkg/cmd/account/payment-methods"
	paymentmethodscascading "github.com/matsuzj/zuora-cli/pkg/cmd/account/payment-methods-cascading"
	paymentmethodsdefault "github.com/matsuzj/zuora-cli/pkg/cmd/account/payment-methods-default"
	"github.com/matsuzj/zuora-cli/pkg/cmd/account/summary"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdAccount creates the account parent command.
func NewCmdAccount(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account <command>",
		Short: "Manage Zuora accounts",
		Long:  "List, view, and inspect Zuora billing accounts.",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(summary.NewCmdSummary(f))
	cmd.AddCommand(paymentmethods.NewCmdPaymentMethods(f))
	cmd.AddCommand(paymentmethodsdefault.NewCmdPaymentMethodsDefault(f))
	cmd.AddCommand(paymentmethodscascading.NewCmdPaymentMethodsCascading(f))

	return cmd
}
