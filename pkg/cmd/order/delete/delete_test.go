package delete

import (
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdDelete(f) }

func TestOrderDelete_Success(t *testing.T) {
	handler := cmdtest.Status(t, "DELETE", "/v1/orders/O-00000001", 204, nil)

	_, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "delete", "O-00000001", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stderr, "Order O-00000001 deleted.")
}

func TestOrderDelete_RequiresConfirm(t *testing.T) {
	cmdtest.RequiresConfirm(t, "order", newCmd, "order", "delete", "O-00000001")
}

func TestOrderDelete_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "delete", "--confirm")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestOrderDelete_BodyResponse(t *testing.T) {
	handler := cmdtest.OK(t, "DELETE", "/v1/orders/O-1", map[string]interface{}{
		"success": true, "jobId": "job-1",
	})

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "delete", "O-1", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "job-1")
}

func TestOrderDelete_UnparseableBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	})

	_, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "delete", "O-1", "--confirm")
	require.NoError(t, err, "non-JSON 200 is a completed delete under the unified policy")
	assert.Contains(t, stderr, "deleted")
}
