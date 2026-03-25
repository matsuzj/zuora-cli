// Package purchaseoptions implements the "zr plan purchase-options" command.
package purchaseoptions

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type purchaseOptionsOpts struct {
	Factory *factory.Factory
	Plan    string
}

// NewCmdPurchaseOptions creates the plan purchase-options command.
func NewCmdPurchaseOptions(f *factory.Factory) *cobra.Command {
	opts := &purchaseOptionsOpts{Factory: f}

	cmd := &cobra.Command{
		Use:   "purchase-options",
		Short: "List purchase options for a plan",
		Long: `List available purchase options for a Zuora commerce plan.

Examples:
  zr plan purchase-options --plan 402880e...
  zr plan purchase-options --plan 402880e... --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Plan == "" {
				return fmt.Errorf("--plan is required")
			}
			return runPurchaseOptions(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Plan, "plan", "", "Rate plan ID (prp_id)")

	return cmd
}

func runPurchaseOptions(cmd *cobra.Command, opts *purchaseOptionsOpts) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"filters": []map[string]interface{}{
			{
				"field":    "prp_id",
				"operator": "=",
				"value":    map[string]string{"string_value": opts.Plan},
			},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := client.Post("/commerce/purchase-options/list", bytes.NewReader(data), api.WithCheckSuccess())
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
