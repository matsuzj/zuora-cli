// Package metricsbyorder implements the "zr ramp metrics-by-order" command.
package metricsbyorder

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdMetricsByOrder creates the ramp metrics-by-order command.
func NewCmdMetricsByOrder(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics-by-order <order-number>",
		Short: "Get ramp metrics by order",
		Long: `Get ramp metrics for a Zuora order.

Examples:
  zr ramp metrics-by-order O-00000001
  zr ramp metrics-by-order O-00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMetricsByOrder(cmd, f, args[0])
		},
	}
	return cmd
}

func runMetricsByOrder(cmd *cobra.Command, f *factory.Factory, orderNumber string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/orders/%s/ramp-metrics", url.PathEscape(orderNumber)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
