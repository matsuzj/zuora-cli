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

func TestOmnichannelDelete_Success204(t *testing.T) {
	handler := cmdtest.Status(t, "DELETE", "/v1/omni-channel-subscriptions/S-001", 204, nil)

	_, stderr, err := cmdtest.Run(t, "omnichannel", newCmd, handler, "omnichannel", "delete", "S-001", "--confirm")

	require.NoError(t, err)
	assert.Contains(t, stderr, "Omni-channel subscription S-001 deleted.")
}

func TestOmnichannelDelete_RequiresConfirm(t *testing.T) {
	cmdtest.RequiresConfirm(t, "omnichannel", newCmd, "omnichannel", "delete", "S-001")
}

func TestOmnichannelDelete_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "omnichannel", newCmd, nil, "omnichannel", "delete", "--confirm")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestOmnichannelDelete_JSON(t *testing.T) {
	handler := cmdtest.Status(t, "", "", 204, nil)

	stdout, _, err := cmdtest.Run(t, "omnichannel", newCmd, handler, "omnichannel", "delete", "S-1", "--confirm", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"success": true`)
}

func TestOmnichannelDelete_BodyResponse(t *testing.T) {
	handler := cmdtest.Status(t, "", "", http.StatusOK, map[string]interface{}{"success": true})

	stdout, _, err := cmdtest.Run(t, "omnichannel", newCmd, handler, "omnichannel", "delete", "S-1", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	// Label-bound (#483): the delete detail view's only row is Success — a bare
	// Contains "true" would pass on any stray "true" anywhere in the output.
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
}

func TestOmnichannelDelete_NonJSONBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	})

	_, stderr, err := cmdtest.Run(t, "omnichannel", newCmd, handler, "omnichannel", "delete", "S-1", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stderr, "Omni-channel subscription S-1 deleted.")
}

func TestOmnichannelDelete_UnparseableBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[1,2,3]`))
	})

	_, _, err := cmdtest.Run(t, "omnichannel", newCmd, handler, "omnichannel", "delete", "S-1", "--confirm")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing response")
}
