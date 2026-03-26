// Package ramp implements the "zr ramp" command group.
package ramp

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/ramp/get"
	getbysub "github.com/matsuzj/zuora-cli/pkg/cmd/ramp/get-by-subscription"
	"github.com/matsuzj/zuora-cli/pkg/cmd/ramp/metrics"
	metricsbyorder "github.com/matsuzj/zuora-cli/pkg/cmd/ramp/metrics-by-order"
	metricsbysub "github.com/matsuzj/zuora-cli/pkg/cmd/ramp/metrics-by-subscription"
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
	cmd.AddCommand(getbysub.NewCmdGetBySubscription(f))
	cmd.AddCommand(metrics.NewCmdMetrics(f))
	cmd.AddCommand(metricsbysub.NewCmdMetricsBySubscription(f))
	cmd.AddCommand(metricsbyorder.NewCmdMetricsByOrder(f))

	return cmd
}
