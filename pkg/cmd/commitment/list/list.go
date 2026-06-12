// Package list implements the "zr commitment list" command.
package list

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type listOptions struct {
	Factory *factory.Factory
	Account string
}

// NewCmdList creates the commitment list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &listOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List commitments for an account",
		Long:  `List commitments associated with a Zuora account.`,
		Example: `  zr commitment list --account-number A00000001
  zr commitment list --account-number A00000001 --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Account == "" { // enforced on the value: the deprecated --account alias must satisfy it
				return fmt.Errorf("--account-number is required")
			}
			return runList(cmd, opts)
		},
	}

	cmdutil.AddAccountNumberFlag(cmd, &opts.Account, "account")

	return cmd
}

func runList(cmd *cobra.Command, opts *listOptions) error {
	f := opts.Factory
	fmtOpts := output.FromCmd(cmd)
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get("/v1/commitments",
		api.WithQuery("accountNumber", opts.Account),
	)
	if err != nil {
		return err
	}

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
