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

func TestPlanCreate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/commerce/plans", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "plan-001",
			"name": "Monthly Plan",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "plan", newCmd, handler, "plan", "create", "--body", `{"name":"Monthly Plan","product_id":"prod-001"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "plan-001")
	assert.Contains(t, stdout, "Monthly Plan")
	assert.Contains(t, stderr, "Plan created.")
}

func TestPlanCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "plan", newCmd, nil, "plan", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestPlanCreate_BareCSVRejectedBeforePost(t *testing.T) {
	// --csv on a JSON-only write must be rejected BEFORE any HTTP call — a
	// rejected-then-retried create could otherwise double-create. nil handler =
	// unexpected requests fail loudly; surfacing the CSV error (not a connection
	// error) proves no POST was attempted.
	_, _, err := cmdtest.Run(t, "plan", newCmd, nil, "plan", "create", "--body", `{"name":"X"}`, "--csv")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--csv is not supported for JSON-only output")
}
