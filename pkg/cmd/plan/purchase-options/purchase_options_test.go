package purchaseoptions

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdPurchaseOptions(f) }

func TestPlanPurchaseOptions_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/commerce/purchase-options/list", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var payload map[string]interface{}
		err = json.Unmarshal(body, &payload)
		require.NoError(t, err)

		filters, ok := payload["filters"].([]interface{})
		require.True(t, ok, "filters should be an array")
		require.Len(t, filters, 1)

		filter := filters[0].(map[string]interface{})
		assert.Equal(t, "prp_id", filter["field"])
		assert.Equal(t, "=", filter["operator"])

		value := filter["value"].(map[string]interface{})
		assert.Equal(t, "plan-123", value["string_value"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": []interface{}{
				map[string]interface{}{
					"id":   "po-9b1c6e42",
					"name": "Purchase Option Fixture #483",
				},
			},
		})
	})

	stdout, _, err := cmdtest.Run(t, "plan", newCmd, handler, "plan", "purchase-options", "--plan", "plan-123")
	require.NoError(t, err)
	assert.Contains(t, stdout, "success")
	// Distinctive row VALUES must be rendered (#483): the old empty-data
	// fixture plus a bare Contains("success") passed for ANY non-crash
	// rendering. Commerce fixtures are not live-verifiable; key shapes kept.
	assert.Contains(t, stdout, "po-9b1c6e42")
	assert.Contains(t, stdout, "Purchase Option Fixture #483")
}

func TestPlanPurchaseOptions_RequiresPlan(t *testing.T) {
	_, _, err := cmdtest.Run(t, "plan", newCmd, nil, "plan", "purchase-options")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "plan" not set`)
}
