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

func TestUsageDelete_Success(t *testing.T) {
	handler := cmdtest.Status(t, "DELETE", "/v1/object/usage/2c92a0f96bd", 204, nil)

	_, stderr, err := cmdtest.Run(t, "usage", newCmd, handler, "usage", "delete", "2c92a0f96bd", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stderr, "Usage record 2c92a0f96bd deleted.")
}

func TestUsageDelete_RequiresConfirm(t *testing.T) {
	_, _, err := cmdtest.Run(t, "usage", newCmd, nil, "usage", "delete", "2c92a0f96bd")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestUsageDelete_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "usage", newCmd, nil, "usage", "delete", "--confirm")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestUsageDelete_BodyResponse(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{"Id": "u-1", "Success": true})

	stdout, stderr, err := cmdtest.Run(t, "usage", newCmd, handler, "usage", "delete", "u-1", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "u-1")
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout) // bool Success rendered via GetString (%v)
	assert.Contains(t, stderr, "Usage record u-1 deleted.")
}

func TestUsageDelete_UnparseableBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	})

	_, stderr, err := cmdtest.Run(t, "usage", newCmd, handler, "usage", "delete", "u-1", "--confirm")
	require.NoError(t, err, "non-JSON 200 is a completed delete under the unified policy")
	assert.Contains(t, stderr, "deleted")
}
