// Package charge implements the "zr charge" command group.
package charge

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/charge/create"
	"github.com/matsuzj/zuora-cli/pkg/cmd/charge/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/charge/update"
	updatetiers "github.com/matsuzj/zuora-cli/pkg/cmd/charge/update-tiers"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdCharge creates the charge parent command.
func NewCmdCharge(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "charge <command>",
		Short: "Manage Zuora commerce charges",
		Long:  "Create, update, get, and manage Zuora commerce charges.",
	}

	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(updatetiers.NewCmdUpdateTiers(f))

	return cmd
}
