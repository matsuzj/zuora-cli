// Package changelogbyorder implements the "zr subscription changelog-by-order" command.
package changelogbyorder

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdChangelogByOrder creates the subscription changelog-by-order command.
func NewCmdChangelogByOrder(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "changelog-by-order <order-number>",
		Short: "Get subscription change logs by order",
		Long: `Get subscription change logs filtered by order number.

Examples:
  zr subscription changelog-by-order O-00000001
  zr subscription changelog-by-order O-00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChangelogByOrder(cmd, f, args[0])
		},
	}
	return cmd
}

func runChangelogByOrder(cmd *cobra.Command, f *factory.Factory, orderNumber string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(
		fmt.Sprintf("/v1/subscription-change-logs/orders/%s", url.PathEscape(orderNumber)),
		api.WithCheckSuccess(),
	)
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
