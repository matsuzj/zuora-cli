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

func TestOrderActionUpdate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/orderActions/OA-123", r.URL.Path)

		bodyBytes, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		require.NoError(t, json.Unmarshal(bodyBytes, &reqBody))
		assert.Equal(t, "Active", reqBody["status"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "order-action", newCmd, handler, "order-action", "update", "OA-123", "--body", `{"status":"Active"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	assert.Contains(t, stderr, "Order action OA-123 updated.")
}

func TestOrderActionUpdate_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Invalid order action")

	_, _, err := cmdtest.Run(t, "order-action", newCmd, handler, "order-action", "update", "OA-123", "--body", `{"bad":"data"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid order action")
}

func TestOrderActionUpdate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order-action", newCmd, nil, "order-action", "update", "OA-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
