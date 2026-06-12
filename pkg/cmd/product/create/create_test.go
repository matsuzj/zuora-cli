package create

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCreate(f) }

func TestProductCreate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/commerce/products", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "prod-001",
			"name": "My Product",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "product", newCmd, handler, "product", "create", "--body", `{"name":"My Product"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "prod-001")
	assert.Contains(t, stdout, "My Product")
	assert.Contains(t, stderr, "Product created.")
}

func TestProductCreate_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Missing required field")

	_, _, err := cmdtest.Run(t, "product", newCmd, handler, "product", "create", "--body", `{"bad":"data"}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Missing required field")
}

func TestProductCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "product", newCmd, nil, "product", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
