package cancel

import (
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCancel(f) }

func TestCancel_RequiresConfirm(t *testing.T) {
	_, _, err := cmdtest.Run(t, "data-query", newCmd, nil, "data-query", "cancel", "job-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestCancel_Confirmed(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/query/jobs/job-1", r.URL.Path)
		w.WriteHeader(200)
		w.Write([]byte(`{"data":{"id":"job-1","queryStatus":"cancelled"}}`))
	})
	_, stderr, err := cmdtest.Run(t, "data-query", newCmd, handler, "data-query", "cancel", "job-1", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stderr, "cancelled")
}

func TestCancel_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "data-query", newCmd, nil, "data-query", "cancel")
	require.Error(t, err)
}
