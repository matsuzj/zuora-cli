// Package metrics implements the "zr subscription metrics" command.
package metrics

import (
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type metricsOptions struct {
	Factory             *factory.Factory
	SubscriptionNumbers []string
}

// NewCmdMetrics creates the subscription metrics command.
func NewCmdMetrics(f *factory.Factory) *cobra.Command {
	opts := &metricsOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Get subscription metrics (MRR, TCV, TCB)",
		Long: `Get subscription metrics including MRR, TCV, and TCB.

Examples:
  zr subscription metrics --subscription-numbers A-S001
  zr sub metrics --subscription-numbers A-S001,A-S002 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMetrics(cmd, opts)
		},
	}

	cmd.Flags().StringSliceVar(&opts.SubscriptionNumbers, "subscription-numbers", nil, "Subscription numbers (required, comma-separated)")
	_ = cmd.MarkFlagRequired("subscription-numbers")

	return cmd
}

func runMetrics(cmd *cobra.Command, opts *metricsOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	reqOpts := []api.RequestOption{
		api.WithQuerySlice("subscriptionNumbers[]", opts.SubscriptionNumbers),
	}

	resp, err := client.Get("/v1/subscriptions/subscription-metrics", reqOpts...)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var body struct {
		Metrics []struct {
			SubscriptionNumber string  `json:"subscriptionNumber"`
			MRR                float64 `json:"mrr"`
			TCV                float64 `json:"tcv"`
			TCB                float64 `json:"tcb"`
			Currency           string  `json:"currency"`
		} `json:"subscriptionMetrics"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	cols := []output.Column{
		{Header: "SUBSCRIPTION", Field: "subscriptionNumber"},
		{Header: "MRR", Field: "mrr"},
		{Header: "TCV", Field: "tcv"},
		{Header: "TCB", Field: "tcb"},
		{Header: "CURRENCY", Field: "currency"},
	}

	rows := make([][]string, len(body.Metrics))
	for i, m := range body.Metrics {
		rows[i] = []string{
			m.SubscriptionNumber,
			fmt.Sprintf("%.2f", m.MRR),
			fmt.Sprintf("%.2f", m.TCV),
			fmt.Sprintf("%.2f", m.TCB),
			m.Currency,
		}
	}

	return output.Render(f.IOStreams, resp.Body, fmtOpts, rows, cols)
}
