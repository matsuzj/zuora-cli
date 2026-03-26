// Package list implements the "zr commitment list" command.
package list

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
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
		Long: `List commitments associated with a Zuora account.

Examples:
  zr commitment list --account A00000001
  zr commitment list --account A00000001 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Account == "" {
				return fmt.Errorf("--account is required")
			}
			return runList(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Account, "account", "", "Account number (required)")

	return cmd
}

func runList(cmd *cobra.Command, opts *listOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get("/v1/commitments",
		api.WithQuery("accountNumber", opts.Account),
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
