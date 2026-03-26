// Package metricsbysubscription implements the "zr ramp metrics-by-subscription" command.
package metricsbysubscription

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdMetricsBySubscription creates the ramp metrics-by-subscription command.
func NewCmdMetricsBySubscription(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics-by-subscription <subscription-key>",
		Short: "Get ramp metrics by subscription",
		Long: `Get ramp metrics for a Zuora subscription.

Examples:
  zr ramp metrics-by-subscription A-S00000001
  zr ramp metrics-by-subscription A-S00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMetricsBySubscription(cmd, f, args[0])
		},
	}
	return cmd
}

func runMetricsBySubscription(cmd *cobra.Command, f *factory.Factory, subscriptionKey string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/subscriptions/%s/ramp-metrics", url.PathEscape(subscriptionKey)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	if fmtOpts.JQ != "" {
		return output.PrintJSON(f.IOStreams, resp.Body, fmtOpts.JQ)
	}
	if fmtOpts.Template != "" {
		return output.PrintTemplate(f.IOStreams, resp.Body, fmtOpts.Template)
	}
	return output.PrintJSON(f.IOStreams, resp.Body, "")
}
