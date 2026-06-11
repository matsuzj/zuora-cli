// Package updatetiers implements the "zr charge update-tiers" command.
package updatetiers

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type updateTiersOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdUpdateTiers creates the charge update-tiers command.
func NewCmdUpdateTiers(f *factory.Factory) *cobra.Command {
	opts := &updateTiersOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "update-tiers",
		Short: "Update commerce charge tiers",
		Long: `Update pricing tiers for a Zuora commerce charge.

Examples:
  zr charge update-tiers --body @tiers.json
  zr charge update-tiers --body '{"charge_id":"...","tiers":[...]}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runUpdateTiers(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runUpdateTiers(cmd *cobra.Command, opts *updateTiersOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Put("/commerce/tiers", bodyReader)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	if fmtOpts.JQ != "" || fmtOpts.Template != "" {
		_, err := output.RenderJSON(f.IOStreams, resp.Body, fmtOpts)
		return err
	}

	if err := output.PrintJSON(f.IOStreams, resp.Body, ""); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Tiers updated.\n")
	return nil
}
