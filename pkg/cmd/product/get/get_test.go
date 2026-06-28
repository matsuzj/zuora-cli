package get

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdGet(f) }

func TestProductGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/commerce/products/PROD-001", map[string]interface{}{
		"id":          "prod-001",
		"name":        "My Product",
		"sku":         "SKU-001",
		"description": "A test product",
	})

	stdout, _, err := cmdtest.Run(t, "product", newCmd, handler, "product", "get", "PROD-001")
	require.NoError(t, err)
	// Label-bound (F-08): values under their own labels.
	assert.Regexp(t, `(?m)^Name:\s+My Product$`, stdout)
	assert.Regexp(t, `(?m)^ID:\s+prod-001$`, stdout)
}

func TestProductGet_PathEscape(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/commerce/products/a%2Fb", r.URL.RawPath)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": "a/b"})
	})

	_, _, err := cmdtest.Run(t, "product", newCmd, handler, "product", "get", "a/b")
	require.NoError(t, err)
}

func TestProductGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "product", newCmd, nil, "product", "get")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
