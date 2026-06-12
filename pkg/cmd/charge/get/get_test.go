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
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "charge-001",
			"name": "Monthly Charge",
		})
	})

	stdout, _, err := cmdtest.Run(t, "charge", newCmd, handler, "charge", "get", "CK-001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "charge-001")
	assert.Contains(t, stdout, "Monthly Charge")
}

func TestChargeGet_RequiresKey(t *testing.T) {
	_, _, err := cmdtest.Run(t, "charge", newCmd, nil, "charge", "get")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

// TestChargeGet_DeprecatedKeyFlagStillWorks pins the P5-3c deprecation
// contract: --key keeps working through v0.5.x (removed in v0.6.0).
func TestChargeGet_DeprecatedKeyFlagStillWorks(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "", map[string]interface{}{"success": true})
	_, _, err := cmdtest.Run(t, "charge", newCmd, handler, "charge", "get", "--key", "CK-001")
	require.NoError(t, err)
}
