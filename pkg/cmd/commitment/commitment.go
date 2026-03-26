// Package commitment implements the "zr commitment" command group.
package commitment

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/commitment/balance"
	"github.com/matsuzj/zuora-cli/pkg/cmd/commitment/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/commitment/list"
	"github.com/matsuzj/zuora-cli/pkg/cmd/commitment/periods"
	"github.com/matsuzj/zuora-cli/pkg/cmd/commitment/schedules"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdCommitment creates the commitment parent command.
func NewCmdCommitment(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commitment <command>",
		Short: "Manage Zuora commitments",
		Long:  "List, view, and inspect Zuora commitments.",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(periods.NewCmdPeriods(f))
	cmd.AddCommand(balance.NewCmdBalance(f))
	cmd.AddCommand(schedules.NewCmdSchedules(f))

	return cmd
}
