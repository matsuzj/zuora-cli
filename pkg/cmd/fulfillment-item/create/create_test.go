package create

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCreate(f) }

func TestFulfillmentItemCreate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/fulfillment-items", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "F-001", payload["fulfillmentKey"])

		w.WriteHeader(200)
		// Real shape: bulk endpoint returns created ids under a "fulfillmentItems"
		// array, not a flat top-level "id".
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"fulfillmentItems": []map[string]interface{}{
				{"id": "fi-00000001"},
			},
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "fulfillment-item", newCmd, handler, "fulfillment-item", "create", "--body", `{"fulfillmentKey":"F-001","quantity":5}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "fi-00000001")
	assert.Contains(t, stderr, "Fulfillment item fi-00000001 created.")
}

func TestFulfillmentItemCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, nil, "fulfillment-item", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}

func TestFulfillmentItemCreate_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Invalid fulfillment item data")

	_, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, handler, "fulfillment-item", "create", "--body", `{"bad":"data"}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid fulfillment item data")
}
