package previewasync

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdPreviewAsync(f) }

func TestOrderPreviewAsync_Success(t *testing.T) {
	var gotBody map[string]interface{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/async/orders/preview", r.URL.Path)
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"jobId":   "JOB-00000001",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "preview-async", "--body", `{"orderDate":"2024-01-01"}`)

	require.NoError(t, err)
	assert.Equal(t, "2024-01-01", gotBody["orderDate"])
	assert.Contains(t, stdout, "JOB-00000001")
	assert.Contains(t, stderr, "Async order preview started. Job ID: JOB-00000001")
}

func TestOrderPreviewAsync_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "preview-async")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
