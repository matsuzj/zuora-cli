// Package changelog implements the "zr subscription changelog" command.
package changelog

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type changelogOptions struct {
	Order   string
	Version string
}

// NewCmdChangelog creates the subscription changelog command. It unifies the
// three change-log lookups that were separate commands (#454): the base
// subscription log, the by-order log (--order), and a specific version
// (--version). The old `changelog-by-order` / `changelog-version` commands stay
// as deprecated commands.
func NewCmdChangelog(f *factory.Factory) *cobra.Command {
	opts := &changelogOptions{}

	cmd := &cobra.Command{
		Use:   "changelog [<subscription-number>]",
		Short: "Get subscription change logs",
		Long: `Get change logs for a Zuora subscription.

Look them up three ways (mutually exclusive):
  - by subscription NUMBER (e.g. A-S00000001):  changelog <subscription-number>
  - for a specific version:                     changelog <subscription-number> --version N
  - by the order that changed it:               changelog --order <order-number>`,
		Example: `  zr subscription changelog S-00000001
  zr subscription changelog S-00000001 --version 2
  zr subscription changelog --order O-00000001`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var subscriptionNumber string
			if len(args) == 1 {
				subscriptionNumber = args[0]
			}
			return runChangelog(cmd, f, opts, subscriptionNumber)
		},
	}

	cmd.Flags().StringVar(&opts.Order, "order", "", "Look up change logs by order number instead of subscription")
	cmd.Flags().StringVar(&opts.Version, "version", "", "Restrict to a specific subscription version (requires <subscription-number>)")

	return cmd
}

func runChangelog(cmd *cobra.Command, f *factory.Factory, opts *changelogOptions, subscriptionNumber string) error {
	// Resolve the lookup mode and reject incompatible combinations. --order is a
	// different endpoint keyed by order, so it cannot combine with a
	// subscription-number arg or --version.
	var path string
	switch {
	case opts.Order != "":
		if subscriptionNumber != "" || opts.Version != "" {
			return fmt.Errorf("--order cannot be combined with a <subscription-number> argument or --version")
		}
		path = fmt.Sprintf("/v1/subscription-change-logs/orders/%s", url.PathEscape(opts.Order))
	case subscriptionNumber != "":
		if opts.Version != "" {
			path = fmt.Sprintf("/v1/subscription-change-logs/%s/versions/%s",
				url.PathEscape(subscriptionNumber), url.PathEscape(opts.Version))
		} else {
			path = fmt.Sprintf("/v1/subscription-change-logs/%s", url.PathEscape(subscriptionNumber))
		}
	default:
		return fmt.Errorf("a <subscription-number> argument or --order is required")
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
