package bulkupdate

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdBulkUpdate(f) }

func TestOrderLineItemBulkUpdate_Success(t *testing.T) {
	var gotBody map[string]interface{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/order-line-items/bulk", r.URL.Path)
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"orderLineItems": []map[string]interface{}{
				{"id": "oli-1", "quantity": 5},
			},
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "order-line-item", newCmd, handler,
		"order-line-item", "bulk-update", "--body", `{"orderLineItems":[{"id":"oli-1","quantity":5}]}`)

	require.NoError(t, err)
	require.NotNil(t, gotBody, "request body should be valid JSON")
	assert.Contains(t, gotBody, "orderLineItems")
	assert.Contains(t, stdout, "true")
	// Label-bound (#483): the command's only detail row is Success, read from
	// the response "success" key — a bare Contains "true" would pass on any
	// stray "true" anywhere in the output.
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
	assert.Contains(t, stderr, "Order line items bulk updated.")
}

func TestOrderLineItemBulkUpdate_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Missing required field")

	_, _, err := cmdtest.Run(t, "order-line-item", newCmd, handler,
		"order-line-item", "bulk-update", "--body", `{"orderLineItems":[]}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Missing required field")
}

func TestOrderLineItemBulkUpdate_RequiresBody(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(200)
		_, _ = io.WriteString(w, `{"success": true}`)
	})

	_, _, err := cmdtest.Run(t, "order-line-item", newCmd, handler,
		"order-line-item", "bulk-update")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
	assert.False(t, called, "no HTTP call should be made when --body is omitted")
}
