package get

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdGet(f) }

func TestPlanGet_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/commerce/plans/query", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)
		assert.Equal(t, "RPK-001", reqBody["product_rate_plan_key"])
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "plan-001",
			"name": "Monthly Plan",
		})
	})

	stdout, _, err := cmdtest.Run(t, "plan", newCmd, handler, "plan", "get", "RPK-001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "plan-001")
	assert.Contains(t, stdout, "Monthly Plan")
}

func TestPlanGet_RequiresKey(t *testing.T) {
	_, _, err := cmdtest.Run(t, "plan", newCmd, nil, "plan", "get")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

// TestPlanGet_KeyFlagRemoved pins that the deprecated --key flag is gone
// (v0.7.0): the key is a positional argument now, so --key is an unknown flag.
func TestPlanGet_KeyFlagRemoved(t *testing.T) {
	_, _, err := cmdtest.Run(t, "plan", newCmd, nil, "plan", "get", "--key", "RPK-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown flag: --key")
}
