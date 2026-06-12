// Package update implements the "zr fulfillment update" command.
package update

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type updateOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdUpdate creates the fulfillment update command.
func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	opts := &updateOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "update <fulfillment-key>",
		Short: "Update a fulfillment",
		Long:  `Update an existing Zuora fulfillment.`,
		Example: `  zr fulfillment update F-00000001 --body @update.json
  zr fulfillment update F-00000001 --body '{"quantity":10}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd, opts, args[0])
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runUpdate(cmd *cobra.Command, opts *updateOptions, fulfillmentKey string) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/fulfillments/%s", url.PathEscape(fulfillmentKey)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			// PUT /v1/fulfillments/{key} returns only {processId, requestId,
			// success, reasons} — no "key" field (sibling of the #56 class).
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
				{Key: "Process ID", Value: cmdutil.GetString(raw, "processId")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Fulfillment %s updated.\n", fulfillmentKey)
		},
	})
}
