package list

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdList(f) }

func TestPlanList_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/commerce/plans/list", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    []interface{}{},
		})
	})

	stdout, _, err := cmdtest.Run(t, "plan", newCmd, handler, "plan", "list", "--body", `{"page":0,"page_size":20}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "success")
}

func TestPlanList_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "plan", newCmd, nil, "plan", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
