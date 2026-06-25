// Package get implements the "zr account get" command.
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

// NewCmdGet creates the account get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <account-key>",
		Short: "Get account details",
		Long: `Get detailed information about a Zuora billing account.

The account-key can be an account ID or account number.`,
		Example: `  zr account get A00000001
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

	// Currency placement varies (F-18): a full account get carries it under
	// billingAndPayment, but a leaner response may only have it under metrics /
	// metricsData (a real account has it in both). Prefer billingAndPayment, fall
	// back to metrics so the field is not blank for the leaner shape.
	currency := cmdutil.GetString(billing, "currency")
	if currency == "" {
		currency = cmdutil.GetString(metrics, "currency")
	}

	fields := []output.DetailField{
		{Key: "ID", Value: cmdutil.GetString(basicInfo, "id")},
		{Key: "Name", Value: cmdutil.GetString(basicInfo, "name")},
		{Key: "Account Number", Value: cmdutil.GetString(basicInfo, "accountNumber")},
		{Key: "Status", Value: cmdutil.GetString(basicInfo, "status")},
		{Key: "Balance", Value: cmdutil.GetMoney(metrics, "balance")},
		{Key: "Currency", Value: currency},
		{Key: "Auto Pay", Value: cmdutil.GetBool(billing, "autoPay")},
		{Key: "Payment Term", Value: cmdutil.GetString(billing, "paymentTerm")},
		{Key: "Bill Cycle Day", Value: cmdutil.GetInt(billing, "billCycleDay")},
		{Key: "CRM ID", Value: cmdutil.GetString(basicInfo, "crmId")},
		{Key: "Sales Rep", Value: cmdutil.GetString(basicInfo, "salesRep")},
		{Key: "Batch", Value: cmdutil.GetString(basicInfo, "batch")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}
