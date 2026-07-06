// Package ramp implements the "zr ramp" command group.
package ramp

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/ramp/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/ramp/list"
	"github.com/matsuzj/zuora-cli/pkg/cmd/ramp/metrics"
	"github.com/spf13/cobra"
)

// NewCmdRamp creates the ramp parent command.
func NewCmdRamp(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ramp <command>",
		Short: "Manage Zuora ramps",
		Long:  "View ramp details and metrics.",
	}

	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(metrics.NewCmdMetrics(f))

	return cmd
}
