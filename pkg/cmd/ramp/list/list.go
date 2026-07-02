// Package list implements the "zr ramp list" command.
package list

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type listOptions struct {
	Subscription string
}

// NewCmdList creates the ramp list command. The ramps endpoint is always keyed
// by subscription (GET /v1/subscriptions/{key}/ramps) and returns multiple
// ramps, so this is a list, not a get — it replaces the mis-named
// `ramp get-by-subscription` (#454), which stays as a deprecated alias.
func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list --subscription <subscription-key>",
		Short: "List ramps for a subscription",
		Long:  `List the ramps associated with a Zuora subscription.`,
		Example: `  zr ramp list --subscription A-S00000001
  zr ramp list --subscription A-S00000001 --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Subscription, "subscription", "", "Subscription number or key (required)")
	_ = cmd.MarkFlagRequired("subscription")

	return cmd
}

func runList(cmd *cobra.Command, f *factory.Factory, opts *listOptions) error {
	fmtOpts := output.FromCmd(cmd)
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/subscriptions/%s/ramps", url.PathEscape(opts.Subscription)))
	if err != nil {
		return err
	}

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
