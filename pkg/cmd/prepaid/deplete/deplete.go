// Package deplete implements the "zr prepaid deplete" command.
package deplete

import (
	"encoding/json"
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

This irreversibly consumes prepaid balance. Use --confirm to proceed.

Examples:
  zr prepaid deplete --body @deplete.json --confirm
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

	cmd.Flags().StringVarP(&opts.Body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Confirm the depletion")

	return cmd
}

func runDeplete(cmd *cobra.Command, opts *depleteOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post("/v1/prepaid-balance-funds/deplete", bodyReader)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Success", Value: cmdutil.GetString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Prepaid balance depleted.\n")
	return nil
}
