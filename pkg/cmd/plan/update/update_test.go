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

func TestPlanUpdate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/commerce/plans", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		// The --body payload must reach the server intact (#484).
		body, rerr := io.ReadAll(r.Body)
		if assert.NoError(t, rerr) {
			assert.JSONEq(t, `{"id":"plan-001","name":"Updated Plan"}`, string(body))
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "plan-001",
			"name": "Updated Plan",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "plan", newCmd, handler, "plan", "update", "--body", `{"id":"plan-001","name":"Updated Plan"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "plan-001")
	assert.Contains(t, stdout, "Updated Plan")
	assert.Contains(t, stderr, "Plan updated.")
}

func TestPlanUpdate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "plan", newCmd, nil, "plan", "update")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
