// Package get implements the "zr fulfillment get" command.
package get

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the fulfillment get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <fulfillment-key>",
		Short: "Get fulfillment details",
		Long: `Get detailed information about a Zuora fulfillment.

Examples:
  zr fulfillment get F-00000001
  zr fulfillment get F-00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, fulfillmentKey string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/fulfillments/%s", url.PathEscape(fulfillmentKey)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	// GET /v1/fulfillments/{key} nests the fulfillment under a top-level
	// "fulfillment" object; there is no top-level "key" (it is keyed by
	// "id"/"fulfillmentNumber"), and the date field is "fulfillmentDate". Fall
	// back to the top level so an unwrapped response still renders.
	ful, _ := raw["fulfillment"].(map[string]interface{})
	if ful == nil {
		ful = raw
	}
	fields := []output.DetailField{
		{Key: "Fulfillment Number", Value: cmdutil.GetString(ful, "fulfillmentNumber")},
		{Key: "ID", Value: cmdutil.GetString(ful, "id")},
		{Key: "State", Value: cmdutil.GetString(ful, "state")},
		{Key: "Order Line Item ID", Value: cmdutil.GetString(ful, "orderLineItemId")},
		{Key: "Quantity", Value: cmdutil.GetDecimal(ful, "quantity")},
		{Key: "Date", Value: cmdutil.GetString(ful, "fulfillmentDate")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}
