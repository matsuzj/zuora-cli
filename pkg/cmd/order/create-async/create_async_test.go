package createasync

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCreateAsync(f) }

func TestOrderCreateAsync_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/async/orders", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var sent map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&sent))
		assert.Equal(t, "A001", sent["existingAccountNumber"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"jobId":   "job-12345",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "create-async", "--body", `{"existingAccountNumber":"A001"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "job-12345")
	assert.Contains(t, stderr, "Async order creation started. Job ID: job-12345")
}

func TestOrderCreateAsync_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Missing required field")

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "create-async", "--body", `{"bad":"data"}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Missing required field")
}

func TestOrderCreateAsync_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "create-async")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
