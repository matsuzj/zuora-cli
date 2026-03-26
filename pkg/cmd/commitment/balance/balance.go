// Package balance implements the "zr commitment balance" command.
package balance

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdBalance creates the commitment balance command.
func NewCmdBalance(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance <commitment-id>",
		Short: "Get commitment balance",
		Long: `Get the balance of a Zuora commitment.

Examples:
  zr commitment balance 2c92c0f8...
  zr commitment balance 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBalance(cmd, f, args[0])
		},
	}
	return cmd
}

func runBalance(cmd *cobra.Command, f *factory.Factory, commitmentID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/commitments/%s/balance", url.PathEscape(commitmentID)), api.WithCheckSuccess())
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
