// Package summary implements the "zr account summary" command.
package summary

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
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
		{Key: "ID", Value: cmdutil.GetString(basicInfo, "id")},
		{Key: "Name", Value: cmdutil.GetString(basicInfo, "name")},
		{Key: "Account Number", Value: cmdutil.GetString(basicInfo, "accountNumber")},
		{Key: "Status", Value: cmdutil.GetString(basicInfo, "status")},
		{Key: "Balance", Value: cmdutil.GetMoney(basicInfo, "balance")},
		{Key: "Currency", Value: cmdutil.GetString(basicInfo, "currency")},
		{Key: "Default Payment Method", Value: getPaymentMethodSummary(basicInfo)},
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

func getPaymentMethodSummary(basicInfo map[string]interface{}) string {
	pm, ok := basicInfo["defaultPaymentMethod"].(map[string]interface{})
	if !ok || pm == nil {
		return ""
	}
	typ := cmdutil.GetString(pm, "paymentMethodType")
	id := cmdutil.GetString(pm, "id")
	if typ != "" && id != "" {
		return fmt.Sprintf("%s (%s)", typ, id)
	}
	if id != "" {
		return id
	}
	return typ
}
