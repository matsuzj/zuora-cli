// Package get implements the "zr charge get" command.
package get

import (
	"bytes"
	"encoding/json"
	"fmt"

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
		Use:   "get <charge-key>",
		Short: "Get a commerce charge by key",
		Long:  `Get a Zuora commerce charge by querying with a charge key.`,
		Example: `  zr charge get CK-001
  zr charge get CK-001 --json`,
		// One positional key; RangeArgs keeps the deprecated --key form
		// parseable through v0.5.x (removed in v0.6.0).
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				opts.Key = args[0]
			}
			if opts.Key == "" {
				// cobra's ExactArgs(1) wording — what this command becomes
				// once the deprecated --key alias is removed.
				return fmt.Errorf("accepts 1 arg(s), received 0")
			}
			return runGet(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Key, "key", "", "Charge key (product_rate_plan_charge_key)")
	_ = cmd.Flags().MarkDeprecated("key", "pass the key as a positional argument instead")

	return cmd
}

func runGet(cmd *cobra.Command, opts *getOptions) error {
	f := opts.Factory
	fmtOpts := output.FromCmd(cmd)
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	payload := map[string]string{"product_rate_plan_charge_key": opts.Key}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := client.Post("/commerce/charges/query", bytes.NewReader(data))
	if err != nil {
		return err
	}

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
