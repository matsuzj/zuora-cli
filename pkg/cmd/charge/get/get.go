// Package get implements the "zr charge get" command.
package get

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the charge get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <charge-key>",
		Short: "Get a commerce charge by key",
		Long:  `Get a Zuora commerce charge by querying with a charge key.`,
		Example: `  zr charge get CK-001
  zr charge get CK-001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}

	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, key string) error {
	payload, err := json.Marshal(map[string]string{"product_rate_plan_charge_key": key})
	if err != nil {
		return err
	}

	// Doc-verified shape (#453): POST /commerce/charges/query returns the
	// product-rate-plan-charge object bare at top level (no wrapper, no
	// success flag at 200); pricingSummary is an array of display strings.
	// Commerce is unprovisioned on the dev sandbox, so the shape comes from
	// the published API reference, like the #435 batch.
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/commerce/charges/query",
		Body:   bytes.NewReader(payload),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			var pricing string
			if arr, ok := raw["pricingSummary"].([]interface{}); ok {
				parts := make([]string, 0, len(arr))
				for _, v := range arr {
					if s, ok := v.(string); ok {
						parts = append(parts, s)
					}
				}
				pricing = strings.Join(parts, ", ")
			}
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Name", Value: cmdutil.GetString(raw, "name")},
				{Key: "Number", Value: cmdutil.GetString(raw, "productRatePlanChargeNumber")},
				{Key: "Charge Type", Value: cmdutil.GetString(raw, "chargeType")},
				{Key: "Charge Model", Value: cmdutil.GetString(raw, "chargeModel")},
				{Key: "Pricing", Value: pricing},
				{Key: "Unit of Measure", Value: cmdutil.GetString(raw, "unitOfMeasure")},
				{Key: "Trigger Event", Value: cmdutil.GetString(raw, "triggerEvent")},
				{Key: "Tax Mode", Value: cmdutil.GetString(raw, "taxMode")},
			}
		},
	})
}
