package create

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCreate(f) }

func TestProductCreate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/commerce/products", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		// The --body payload must reach the server intact (#484).
		body, rerr := io.ReadAll(r.Body)
		if assert.NoError(t, rerr) {
			assert.JSONEq(t, `{"name":"My Product"}`, string(body))
		}
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

func TestProductCreate_BareCSVRejectedBeforePost(t *testing.T) {
	// --csv on a JSON-only write must be rejected BEFORE any HTTP call — a
	// rejected-then-retried create could otherwise double-create. nil handler =
	// unexpected requests fail loudly; surfacing the CSV error (not a connection
	// error) proves no POST was attempted.
	_, _, err := cmdtest.Run(t, "product", newCmd, nil, "product", "create", "--body", `{"name":"X"}`, "--csv")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--csv is not supported for JSON-only output")
}
