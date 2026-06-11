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

func TestUsageUpdate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/object/usage/usage123", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "Quantity")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Success": true,
			"Id":      "usage123",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "usage", newCmd, handler, "usage", "update", "usage123", "--body", `{"Quantity":20}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "usage123")
	assert.Contains(t, stderr, "Usage record usage123 updated.")
}

func TestUsageUpdate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "usage", newCmd, nil, "usage", "update", "usage123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}

func TestUsageUpdate_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "usage", newCmd, nil, "usage", "update")

	assert.Error(t, err)
}
