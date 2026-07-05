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
	// Doc-verified (#435): the operation is POST, the response is the product
	// object at top level with camelCase keys (startDate/endDate — the old
	// snake_case keys never existed) and no top-level description.
	handler := cmdtest.OK(t, "POST", "/commerce/products/PROD-001", map[string]interface{}{
		"id":        "prod-001",
		"name":      "My Product",
		"sku":       "SKU-001",
		"state":     "Active",
		"startDate": "2026-02-03",
		"endDate":   "2031-04-05",
	})

	stdout, _, err := cmdtest.Run(t, "product", newCmd, handler, "product", "get", "PROD-001")
	require.NoError(t, err)
	// Label-bound (F-08): values under their own labels.
	assert.Regexp(t, `(?m)^Name:\s+My Product$`, stdout)
	assert.Regexp(t, `(?m)^ID:\s+prod-001$`, stdout)
	assert.Regexp(t, `(?m)^SKU:\s+SKU-001$`, stdout)
	assert.Regexp(t, `(?m)^State:\s+Active$`, stdout)
	assert.Regexp(t, `(?m)^Start Date:\s+2026-02-03$`, stdout)
	assert.Regexp(t, `(?m)^End Date:\s+2031-04-05$`, stdout)
}

// TestProductGet_AllowedInReadOnlyMode pins the read-only allowlist entry for
// the POST-that-is-a-read retrieve operation (#435): without the
// ^commerce/products/[^/]+$ pattern, --read-only would block a pure read.
func TestProductGet_AllowedInReadOnlyMode(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "/commerce/products/PROD-001", map[string]interface{}{
		"id": "prod-001", "name": "RO Product",
	})

	stdout, _, err := cmdtest.Run(t, "product", newCmd, handler, "product", "get", "PROD-001", "--read-only")
	require.NoError(t, err, "product get is a read and must pass in read-only mode despite using POST")
	assert.Regexp(t, `(?m)^Name:\s+RO Product$`, stdout)
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
