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

func TestGet_DescendsIntoData(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/query/jobs/job-1", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{
			"id": "job-1", "queryStatus": "completed", "outputRows": "42", "dataFile": "https://s3/x",
		}})
	})
	stdout, _, err := cmdtest.Run(t, "data-query", newCmd, handler, "data-query", "get", "job-1")
	require.NoError(t, err)
	// Label-bound (F-08). Data File is the PRIMARY output of dq get/submit — a
	// wrong/missing key would render it blank yet pass a bare Contains. (#432)
	assert.Regexp(t, `(?m)^Status:\s+completed$`, stdout)
	assert.Regexp(t, `(?m)^Output Rows:\s+42$`, stdout)
	assert.Regexp(t, `(?m)^Data File:\s+https://s3/x$`, stdout)
}

func TestGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "data-query", newCmd, nil, "data-query", "get")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
