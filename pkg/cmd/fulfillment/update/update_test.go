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

func TestFulfillmentUpdate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/fulfillments/F-00000001", r.URL.Path)

		bodyBytes, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		require.NoError(t, json.Unmarshal(bodyBytes, &reqBody))
		assert.Equal(t, float64(10), reqBody["quantity"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"key":     "F-00000001",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "fulfillment", newCmd, handler, "fulfillment", "update", "F-00000001", "--body", `{"quantity":10}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "F-00000001")
	assert.Contains(t, stderr, "Fulfillment F-00000001 updated.")
}

func TestFulfillmentUpdate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "fulfillment", newCmd, nil, "fulfillment", "update", "F-00000001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestFulfillmentUpdate_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Invalid fulfillment data")

	_, _, err := cmdtest.Run(t, "fulfillment", newCmd, handler, "fulfillment", "update", "F-00000001", "--body", `{"bad":"data"}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid fulfillment data")
}
