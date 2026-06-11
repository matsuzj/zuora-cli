// Package create implements the "zr fulfillment create" command.
package create

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type createOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdCreate creates the fulfillment create command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &createOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a fulfillment",
		Long: `Create a new Zuora fulfillment.

Examples:
  zr fulfillment create --body @fulfillment.json
  zr fulfillment create --body '{"orderLineItemId":"..."}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runCreate(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runCreate(cmd *cobra.Command, opts *createOptions) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/fulfillments",
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			fulfillmentID, fulfillmentNumber := firstFulfillment(raw)
			return []output.DetailField{
				{Key: "Fulfillment Number", Value: fulfillmentNumber},
				{Key: "ID", Value: fulfillmentID},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			fulfillmentID, fulfillmentNumber := firstFulfillment(raw)
			ref := fulfillmentNumber
			if ref == "" {
				ref = fulfillmentID
			}
			if ref != "" {
				return fmt.Sprintf("Fulfillment %s created.\n", ref)
			}
			return ""
		},
	})
}

// firstFulfillment extracts the first created fulfillment's id and number from
// the bulk response. POST /v1/fulfillments is the BULK create endpoint: the
// created object(s) are returned under a top-level "fulfillments" array (keyed
// by id/fulfillmentNumber), not a flat top-level "key". "success" is top-level.
func firstFulfillment(raw map[string]interface{}) (id, number string) {
	if arr, ok := raw["fulfillments"].([]interface{}); ok && len(arr) > 0 {
		if first, ok := arr[0].(map[string]interface{}); ok {
			id = cmdutil.GetString(first, "id")
			number = cmdutil.GetString(first, "fulfillmentNumber")
		}
	}
	return id, number
}
