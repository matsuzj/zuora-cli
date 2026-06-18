// Package get implements the "zr plan get" command.
package get

import (
	"bytes"
	"encoding/json"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type getOptions struct {
	Factory *factory.Factory
	Key     string
}

// NewCmdGet creates the plan get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	opts := &getOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "get <rate-plan-key>",
		Short: "Get a commerce plan by key",
		Long:  `Get a Zuora commerce plan by querying with a rate plan key.`,
		Example: `  zr plan get RPK-001
  zr plan get RPK-001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Key = args[0]
			return runGet(cmd, opts)
		},
	}

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

	payload := map[string]string{"product_rate_plan_key": opts.Key}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := client.Post("/commerce/plans/query", bytes.NewReader(data))
	if err != nil {
		return err
	}

	return output.RenderJSONOnly(f.IOStreams, resp.Body, fmtOpts)
}
