package preview

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdPreview(f) }

func TestOrderPreview_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/orders/preview", r.URL.Path)
		var body map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "2024-01-01", body["orderDate"])
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"previewResult": map[string]interface{}{
				"charges": []map[string]interface{}{{"number": "C-001"}},
			},
		})
	})

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "preview", "--body", `{"orderDate":"2024-01-01"}`)
	require.NoError(t, err)
	// preview prints the raw JSON response.
	assert.Contains(t, stdout, "previewResult")
	assert.Contains(t, stdout, "C-001")
}

func TestOrderPreview_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Missing required field")

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "preview", "--body", `{"bad":"data"}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Missing required field")
}

func TestOrderPreview_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "preview")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
