// Package prepaid implements the "zr prepaid" command group.
package prepaid

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/prepaid/deplete"
	reverserollover "github.com/matsuzj/zuora-cli/pkg/cmd/prepaid/reverse-rollover"
	"github.com/matsuzj/zuora-cli/pkg/cmd/prepaid/rollover"
	"github.com/spf13/cobra"
)

// NewCmdPrepaid creates the prepaid parent command.
func NewCmdPrepaid(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepaid <command>",
		Short: "Manage prepaid balance operations",
		Long:  "Rollover, reverse rollover, and deplete prepaid balances.",
	}

	cmd.AddCommand(rollover.NewCmdRollover(f))
	cmd.AddCommand(reverserollover.NewCmdReverseRollover(f))
	cmd.AddCommand(deplete.NewCmdDeplete(f))

	return cmd
}
