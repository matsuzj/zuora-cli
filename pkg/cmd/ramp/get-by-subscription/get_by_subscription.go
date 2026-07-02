// Package getbysubscription implements the "zr ramp get-by-subscription" command.
package getbysubscription

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGetBySubscription creates the ramp get-by-subscription command.
func NewCmdGetBySubscription(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:        "get-by-subscription <subscription-key>",
		Short:      "Get ramps by subscription",
		Long:       `Get ramps associated with a Zuora subscription.`,
		Deprecated: `use "ramp list --subscription <subscription-key>" instead`,
		Example: `  zr ramp get-by-subscription A-S00000001
  zr ramp get-by-subscription A-S00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetBySubscription(cmd, f, args[0])
		},
	}
	return cmd
}

func runGetBySubscription(cmd *cobra.Command, f *factory.Factory, subscriptionKey string) error {
	fmtOpts := output.FromCmd(cmd)
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/subscriptions/%s/ramps", url.PathEscape(subscriptionKey)))
	if err != nil {
		return err
	}

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
