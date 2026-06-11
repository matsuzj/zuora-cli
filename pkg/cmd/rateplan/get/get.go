// Package get implements the "zr rateplan get" command.
package get

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the rateplan get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <rate-plan-id>",
		Short: "Get a rate plan",
		Long: `Get detailed information about a Zuora rate plan (v1 API).

Examples:
  zr rateplan get 402880e...
  zr rateplan get 402880e... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, ratePlanID string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/rateplans/%s", url.PathEscape(ratePlanID)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	// GET /v1/rateplans/{id} returns a SUBSCRIPTION rate plan. Its real fields
	// (verified live) are id/ratePlanName/productId/productName/productSku/
	// productRatePlanId/subscriptionId/subscriptionVersion — NOT the product-
	// catalog fields (name/status/description/effective*Date) this previously read.
	fields := []output.DetailField{
		{Key: "ID", Value: cmdutil.GetString(raw, "id")},
		{Key: "Rate Plan Name", Value: cmdutil.GetString(raw, "ratePlanName")},
		{Key: "Product ID", Value: cmdutil.GetString(raw, "productId")},
		{Key: "Product Name", Value: cmdutil.GetString(raw, "productName")},
		{Key: "Product SKU", Value: cmdutil.GetString(raw, "productSku")},
		{Key: "Product Rate Plan ID", Value: cmdutil.GetString(raw, "productRatePlanId")},
		{Key: "Subscription ID", Value: cmdutil.GetString(raw, "subscriptionId")},
		{Key: "Subscription Version", Value: cmdutil.GetString(raw, "subscriptionVersion")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}
