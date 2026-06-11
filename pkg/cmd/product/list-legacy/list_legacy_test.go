package listlegacy

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdListLegacy(f) }

func TestProductListLegacy_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/commerce/legacy/products/list", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    []interface{}{},
		})
	})

	stdout, _, err := cmdtest.Run(t, "product", newCmd, handler, "product", "list-legacy", "--body", `{"page":0,"page_size":20}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "success")
}

func TestProductListLegacy_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "product", newCmd, nil, "product", "list-legacy")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
