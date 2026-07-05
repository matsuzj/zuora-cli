package get

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdGet(f) }

func TestChargeGet_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/commerce/charges/query", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)
		assert.Equal(t, "CK-001", reqBody["product_rate_plan_charge_key"])
		// Doc-verified shape (#453): bare top-level PRPC object, no success
		// flag; pricingSummary is an array of display strings.
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":                          "charge-001",
			"name":                        "Monthly Charge",
			"productRatePlanChargeNumber": "PRPC-00000279",
			"chargeType":                  "recurring",
			"chargeModel":                 "flat_fee",
			"pricingSummary":              []interface{}{"USD100", "JPY9800"},
			"unitOfMeasure":               "License",
			"triggerEvent":                "contract_effective",
			"taxMode":                     "tax_exclusive",
		})
	})

	stdout, _, err := cmdtest.Run(t, "charge", newCmd, handler, "charge", "get", "CK-001")
	require.NoError(t, err)
	// Label-bound (F-08): each value under its own label; Pricing joins the
	// pricingSummary array.
	assert.Regexp(t, `(?m)^ID:\s+charge-001$`, stdout)
	assert.Regexp(t, `(?m)^Name:\s+Monthly Charge$`, stdout)
	assert.Regexp(t, `(?m)^Number:\s+PRPC-00000279$`, stdout)
	assert.Regexp(t, `(?m)^Charge Type:\s+recurring$`, stdout)
	assert.Regexp(t, `(?m)^Charge Model:\s+flat_fee$`, stdout)
	assert.Regexp(t, `(?m)^Pricing:\s+USD100, JPY9800$`, stdout)
	assert.Regexp(t, `(?m)^Unit of Measure:\s+License$`, stdout)
	assert.Regexp(t, `(?m)^Trigger Event:\s+contract_effective$`, stdout)
	assert.Regexp(t, `(?m)^Tax Mode:\s+tax_exclusive$`, stdout)
}

func TestChargeGet_RequiresKey(t *testing.T) {
	_, _, err := cmdtest.Run(t, "charge", newCmd, nil, "charge", "get")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

// TestChargeGet_KeyFlagRemoved pins that the deprecated --key flag is gone
// (v0.7.0): the key is a positional argument now, so --key is an unknown flag.
func TestChargeGet_KeyFlagRemoved(t *testing.T) {
	_, _, err := cmdtest.Run(t, "charge", newCmd, nil, "charge", "get", "--key", "CK-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown flag: --key")
}
