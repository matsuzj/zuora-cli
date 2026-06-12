// Package deplete implements the "zr prepaid deplete" command.
package deplete

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type depleteOptions struct {
	Factory *factory.Factory
	Body    string
	Confirm bool
}

// NewCmdDeplete creates the prepaid deplete command.
func NewCmdDeplete(f *factory.Factory) *cobra.Command {
	opts := &depleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "deplete",
		Short: "Deplete prepaid balance",
		Long: `Deplete a prepaid balance fund in Zuora.

This irreversibly consumes prepaid balance. Use --confirm to proceed.`,
		Example: `  zr prepaid deplete --body @deplete.json --confirm
  zr prepaid deplete --body '{"amount":100,"currency":"USD"}' --confirm`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			if err := cmdutil.RequireConfirm(opts.Confirm); err != nil {
				return err
			}
			return runDeplete(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)
	cmdutil.AddConfirmFlag(cmd, &opts.Confirm, "depletion")

	return cmd
}

func runDeplete(cmd *cobra.Command, opts *depleteOptions) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/prepaid-balance-funds/deplete",
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return "Prepaid balance depleted.\n"
		},
	})
}
