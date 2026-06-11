// Package balance implements the "zr commitment balance" command.
package balance

import (
	"fmt"
	"net/url"

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
	fmtOpts := output.FromCmd(cmd)
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/commitments/%s/balance", url.PathEscape(commitmentID)))
	if err != nil {
		return err
	}

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
