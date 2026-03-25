// Package get implements the "zr rateplan get" command.
package get

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
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

	resp, err := client.Get(fmt.Sprintf("/v1/rateplans/%s", url.PathEscape(ratePlanID)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "ID", Value: getString(raw, "id")},
		{Key: "Name", Value: getString(raw, "name")},
		{Key: "Product ID", Value: getString(raw, "productId")},
		{Key: "Product Name", Value: getString(raw, "productName")},
		{Key: "Product Rate Plan Number", Value: getString(raw, "productRatePlanNumber")},
		{Key: "Status", Value: getString(raw, "status")},
		{Key: "Description", Value: getString(raw, "description")},
		{Key: "Effective Start Date", Value: getString(raw, "effectiveStartDate")},
		{Key: "Effective End Date", Value: getString(raw, "effectiveEndDate")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
