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

func TestOrderUpdate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/orders/O-00000001", r.URL.Path)

		var body map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "Completed", body["orderDate"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"orderNumber": "O-00000001",
			"status":      "Completed",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "update", "O-00000001", "--body", `{"orderDate":"Completed"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000001")
	assert.Contains(t, stderr, "Order O-00000001 updated.")
}

func TestOrderUpdate_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Missing required field")

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "update", "O-00000001", "--body", `{"bad":"data"}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Missing required field")
}

func TestOrderUpdate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "update", "O-00000001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
