// Package meter implements the "zr meter" command group.
package meter

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/meter/audit"
	"github.com/matsuzj/zuora-cli/pkg/cmd/meter/debug"
	"github.com/matsuzj/zuora-cli/pkg/cmd/meter/run"
	"github.com/matsuzj/zuora-cli/pkg/cmd/meter/status"
	"github.com/matsuzj/zuora-cli/pkg/cmd/meter/summary"
	"github.com/spf13/cobra"
)

// NewCmdMeter creates the meter parent command.
func NewCmdMeter(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meter <command>",
		Short: "Manage Zuora usage meters",
		Long:  "Run, debug, check status, summarize, and audit Zuora usage meters.",
	}

	cmd.AddCommand(run.NewCmdRun(f))
	cmd.AddCommand(debug.NewCmdDebug(f))
	cmd.AddCommand(status.NewCmdStatus(f))
	cmd.AddCommand(summary.NewCmdSummary(f))
	cmd.AddCommand(audit.NewCmdAudit(f))

	return cmd
}
