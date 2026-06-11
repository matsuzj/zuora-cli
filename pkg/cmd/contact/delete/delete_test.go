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

func TestContactDelete_Success(t *testing.T) {
	handler := cmdtest.OK(t, "DELETE", "/v1/contacts/c-123", map[string]interface{}{
		"success": true,
	})

	_, stderr, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "delete", "c-123", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stderr, "deleted")
}

func TestContactDelete_RequiresConfirm(t *testing.T) {
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "delete", "c-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestContactDelete_EmptyBodyJSON(t *testing.T) {
	handler := cmdtest.Status(t, "", "", 204, nil)

	stdout, _, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "delete", "c-1", "--confirm", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"success": true`)
}

func TestContactDelete_BodyMissingSuccess(t *testing.T) {
	handler := cmdtest.Status(t, "", "", 200, map[string]interface{}{})

	_, stderr, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "delete", "c-1", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stderr, "deleted")
}

func TestContactDelete_UnparseableBody(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("not json"))
	}

	_, stderr, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "delete", "c-1", "--confirm")
	require.NoError(t, err, "non-JSON 200 is a completed delete under the unified policy")
	assert.Contains(t, stderr, "deleted")
}
