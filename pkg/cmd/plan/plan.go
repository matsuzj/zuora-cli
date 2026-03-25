// Package plan implements the "zr plan" command group.
package plan

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/plan/create"
	"github.com/matsuzj/zuora-cli/pkg/cmd/plan/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/plan/list"
	purchaseoptions "github.com/matsuzj/zuora-cli/pkg/cmd/plan/purchase-options"
	"github.com/matsuzj/zuora-cli/pkg/cmd/plan/update"
	"github.com/spf13/cobra"
)

// NewCmdPlan creates the plan parent command.
func NewCmdPlan(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan <command>",
		Short: "Manage Zuora commerce plans",
		Long:  "Create, update, get, list, and query Zuora commerce plans.",
	}

	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(purchaseoptions.NewCmdPurchaseOptions(f))

	return cmd
}
