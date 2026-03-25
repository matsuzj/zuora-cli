// Package get implements the "zr charge get" command.
package get

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type getOptions struct {
	Factory *factory.Factory
	Key     string
}

// NewCmdGet creates the charge get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	opts := &getOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a commerce charge by key",
		Long: `Get a Zuora commerce charge by querying with a charge key.

Examples:
  zr charge get --key CK-001
  zr charge get --key CK-001 --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Key == "" {
				return fmt.Errorf("--key is required")
			}
			return runGet(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Key, "key", "", "Charge key (product_rate_plan_charge_key)")

	return cmd
}

func runGet(cmd *cobra.Command, opts *getOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	payload := map[string]string{"product_rate_plan_charge_key": opts.Key}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := client.Post("/commerce/charges/query", bytes.NewReader(data), api.WithCheckSuccess())
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
