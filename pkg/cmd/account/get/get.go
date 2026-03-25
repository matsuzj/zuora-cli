// Package get implements the "zr account get" command.
package get

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the account get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <account-key>",
		Short: "Get account details",
		Long: `Get detailed information about a Zuora billing account.

The account-key can be an account ID or account number.

Examples:
  zr account get A00000001
  zr account get 402880ec12345 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, key string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/accounts/%s", url.PathEscape(key)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	// Zuora GET /v1/accounts/{key} returns nested sub-objects:
	// basicInfo, billingAndPayment, metrics, billToContact, soldToContact, taxInfo
	basicInfo, _ := raw["basicInfo"].(map[string]interface{})
	if basicInfo == nil {
		basicInfo = raw
	}
	billing, _ := raw["billingAndPayment"].(map[string]interface{})
	if billing == nil {
		billing = raw
	}
	metrics, _ := raw["metrics"].(map[string]interface{})
	if metrics == nil {
		metrics = raw
	}

	fields := []output.DetailField{
		{Key: "ID", Value: getString(basicInfo, "id")},
		{Key: "Name", Value: getString(basicInfo, "name")},
		{Key: "Account Number", Value: getString(basicInfo, "accountNumber")},
		{Key: "Status", Value: getString(basicInfo, "status")},
		{Key: "Balance", Value: getNumber(metrics, "balance")},
		{Key: "Currency", Value: getString(billing, "currency")},
		{Key: "Auto Pay", Value: getBool(billing, "autoPay")},
		{Key: "Payment Term", Value: getString(billing, "paymentTerm")},
		{Key: "Bill Cycle Day", Value: getInt(billing, "billCycleDay")},
		{Key: "CRM ID", Value: getString(basicInfo, "crmId")},
		{Key: "Sales Rep", Value: getString(basicInfo, "salesRep")},
		{Key: "Batch", Value: getString(basicInfo, "batch")},
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

func getBool(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if b, ok := v.(bool); ok {
			return strconv.FormatBool(b)
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if f, ok := v.(float64); ok {
			return strconv.Itoa(int(f))
		}
	}
	return ""
}
