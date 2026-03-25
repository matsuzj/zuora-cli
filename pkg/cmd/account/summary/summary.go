// Package summary implements the "zr account summary" command.
package summary

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdSummary creates the account summary command.
func NewCmdSummary(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary <account-key>",
		Short: "Get account summary",
		Long: `Get a summary of a Zuora billing account including invoices, payments, and usage.

Examples:
  zr account summary A00000001
  zr account summary A00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSummary(cmd, f, args[0])
		},
	}
	return cmd
}

func runSummary(cmd *cobra.Command, f *factory.Factory, key string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/accounts/%s/summary", url.PathEscape(key)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	// Extract basic info from nested structure
	basicInfo, _ := raw["basicInfo"].(map[string]interface{})
	if basicInfo == nil {
		basicInfo = raw
	}

	fields := []output.DetailField{
		{Key: "ID", Value: getString(basicInfo, "id")},
		{Key: "Name", Value: getString(basicInfo, "name")},
		{Key: "Account Number", Value: getString(basicInfo, "accountNumber")},
		{Key: "Status", Value: getString(basicInfo, "status")},
		{Key: "Balance", Value: getNumber(basicInfo, "balance")},
		{Key: "Currency", Value: getString(basicInfo, "currency")},
		{Key: "Default Payment Method", Value: getString(basicInfo, "defaultPaymentMethod")},
	}

	// Add subscription/invoice counts if available
	if subs, ok := raw["subscriptions"].([]interface{}); ok {
		fields = append(fields, output.DetailField{
			Key: "Subscriptions", Value: fmt.Sprintf("%d", len(subs)),
		})
	}
	if invs, ok := raw["invoices"].([]interface{}); ok {
		fields = append(fields, output.DetailField{
			Key: "Invoices", Value: fmt.Sprintf("%d", len(invs)),
		})
	}
	if pmts, ok := raw["payments"].([]interface{}); ok {
		fields = append(fields, output.DetailField{
			Key: "Payments", Value: fmt.Sprintf("%d", len(pmts)),
		})
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getNumber(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if f, ok := v.(float64); ok {
			return fmt.Sprintf("%.2f", f)
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}
