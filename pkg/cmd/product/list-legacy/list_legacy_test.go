package listlegacy

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdListLegacy(f) }

func TestProductListLegacy_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/commerce/legacy/products/list", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		// The --body payload must reach the server intact (#484): the handler
		// previously ignored r.Body.
		body, rerr := io.ReadAll(r.Body)
		if assert.NoError(t, rerr) {
			assert.JSONEq(t, `{"page":0,"page_size":20}`, string(body))
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": []interface{}{
				map[string]interface{}{
					"id":   "prod-7c5e1b90",
					"name": "Legacy Product Fixture #483",
				},
			},
		})
	})

	stdout, _, err := cmdtest.Run(t, "product", newCmd, handler, "product", "list-legacy", "--body", `{"page":0,"page_size":20}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "success")
	// Distinctive row VALUES must be rendered (#483): the old empty-data
	// fixture plus a bare Contains("success") passed for ANY non-crash
	// rendering. Commerce fixtures are not live-verifiable; key shapes kept.
	assert.Contains(t, stdout, "prod-7c5e1b90")
	assert.Contains(t, stdout, "Legacy Product Fixture #483")
}

// TestProductListLegacy_EmptyData pins the empty-result rendering of this
// JSON-only command: the full envelope is emitted verbatim (there is no table
// empty-state here), asserted structurally rather than via a substring.
func TestProductListLegacy_EmptyData(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "/commerce/legacy/products/list", map[string]interface{}{
		"success": true,
		"data":    []interface{}{},
	})

	stdout, _, err := cmdtest.Run(t, "product", newCmd, handler, "product", "list-legacy", "--body", `{"page":0,"page_size":20}`)
	require.NoError(t, err)
	assert.JSONEq(t, `{"success":true,"data":[]}`, stdout)
}

func TestProductListLegacy_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "product", newCmd, nil, "product", "list-legacy")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
