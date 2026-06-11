package update

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdUpdate(f) }

func TestFulfillmentItemUpdate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/fulfillment-items/item-001", r.URL.Path)

		var body map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, float64(10), body["quantity"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      "item-001",
		})
	})

	stdout, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, handler, "fulfillment-item", "update", "item-001", "--body", `{"quantity":10}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "item-001")
}

func TestFulfillmentItemUpdate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, nil, "fulfillment-item", "update", "item-001")
	assert.Error(t, err)
}

func TestFulfillmentItemUpdate_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, nil, "fulfillment-item", "update", "--body", `{"quantity":10}`)
	assert.Error(t, err)
}
