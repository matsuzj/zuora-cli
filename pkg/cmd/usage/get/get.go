// Package get implements the "zr usage get" command.
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

// NewCmdGet creates the usage get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a usage record",
		Long: `Get a usage record by ID via the CRUD API.

Examples:
  zr usage get 2c92a0f96bd...
  zr usage get 2c92a0f96bd... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, id string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/object/usage/%s", url.PathEscape(id)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "ID", Value: getString(raw, "Id")},
		{Key: "Account ID", Value: getString(raw, "AccountId")},
		{Key: "Quantity", Value: getString(raw, "Quantity")},
		{Key: "Start Date", Value: getString(raw, "StartDateTime")},
		{Key: "End Date", Value: getString(raw, "EndDateTime")},
		{Key: "UOM", Value: getString(raw, "UOM")},
		{Key: "Status", Value: getString(raw, "Status")},
		{Key: "Subscription ID", Value: getString(raw, "SubscriptionId")},
		{Key: "Charge ID", Value: getString(raw, "ChargeId")},
		{Key: "Created Date", Value: getString(raw, "CreatedDate")},
		{Key: "Updated Date", Value: getString(raw, "UpdatedDate")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
