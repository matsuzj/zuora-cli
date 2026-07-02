// Package metrics implements the "zr ramp metrics" command.
package metrics

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type metricsOptions struct {
	Order        string
	Subscription string
}

// NewCmdMetrics creates the ramp metrics command. The same ramp-metrics payload
// is reachable keyed by ramp, order, or subscription; those were three separate
// commands (#454). Fold the order/subscription variants into flags. The old
// `metrics-by-order` / `metrics-by-subscription` stay as deprecated commands.
func NewCmdMetrics(f *factory.Factory) *cobra.Command {
	opts := &metricsOptions{}

	cmd := &cobra.Command{
		Use:   "metrics [<ramp-number>]",
		Short: "Get ramp metrics",
		Long: `Get metrics for a Zuora ramp.

Key the lookup by exactly one of (mutually exclusive):
  - ramp number:   metrics <ramp-number>
  - order number:  metrics --order <order-number>
  - subscription:  metrics --subscription <subscription-key>`,
		Example: `  zr ramp metrics R-00000001
  zr ramp metrics --order O-00000001
  zr ramp metrics --subscription A-S00000001`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var rampNumber string
			if len(args) == 1 {
				rampNumber = args[0]
			}
			return runMetrics(cmd, f, opts, rampNumber)
		},
	}

	cmd.Flags().StringVar(&opts.Order, "order", "", "Look up ramp metrics by order number")
	cmd.Flags().StringVar(&opts.Subscription, "subscription", "", "Look up ramp metrics by subscription key")

	return cmd
}

func runMetrics(cmd *cobra.Command, f *factory.Factory, opts *metricsOptions, rampNumber string) error {
	// Exactly one selector: a ramp-number arg, --order, or --subscription. Each
	// is a distinct endpoint returning the same ramp-metrics payload.
	var path string
	set := 0
	if rampNumber != "" {
		set++
		path = fmt.Sprintf("/v1/ramps/%s/ramp-metrics", url.PathEscape(rampNumber))
	}
	if opts.Order != "" {
		set++
		path = fmt.Sprintf("/v1/orders/%s/ramp-metrics", url.PathEscape(opts.Order))
	}
	if opts.Subscription != "" {
		set++
		path = fmt.Sprintf("/v1/subscriptions/%s/ramp-metrics", url.PathEscape(opts.Subscription))
	}
	switch {
	case set == 0:
		return fmt.Errorf("one of <ramp-number>, --order, or --subscription is required")
	case set > 1:
		return fmt.Errorf("specify only one of <ramp-number>, --order, or --subscription")
	}

	fmtOpts := output.FromCmd(cmd)
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
