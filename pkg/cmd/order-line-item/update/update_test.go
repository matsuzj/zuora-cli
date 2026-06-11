package update

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdUpdate(f) }

func TestOrderLineItemUpdate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/order-line-items/OLI-001", r.URL.Path)

		raw, _ := io.ReadAll(r.Body)
		var got map[string]interface{}
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.Equal(t, "Updated description", got["description"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      "OLI-001",
		})
	})

	_, stderr, err := cmdtest.Run(t, "order-line-item", newCmd, handler, "order-line-item", "update", "OLI-001", "--body", `{"description":"Updated description"}`)

	require.NoError(t, err)
	assert.Contains(t, stderr, "Order line item OLI-001 updated.")
}

func TestOrderLineItemUpdate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order-line-item", newCmd, nil, "order-line-item", "update", "OLI-001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}

func TestOrderLineItemUpdate_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order-line-item", newCmd, nil, "order-line-item", "update", "--body", `{"description":"x"}`)

	assert.Error(t, err)
}
