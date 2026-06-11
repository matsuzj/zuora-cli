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

func TestProductUpdate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/commerce/products", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "prod-001",
			"name": "Updated Product",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "product", newCmd, handler, "product", "update", "--body", `{"id":"prod-001","name":"Updated Product"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "prod-001")
	assert.Contains(t, stdout, "Updated Product")
	assert.Contains(t, stderr, "Product updated.")
}

func TestProductUpdate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "product", newCmd, nil, "product", "update")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
