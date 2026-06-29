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

func TestDelete_EmptyBody(t *testing.T) {
	// A 204 / empty 200 response must not crash with "unexpected end of JSON
	// input" — RenderDeleteResult guards the empty body before parsing. Bites if
	// the handler reverts to a raw json.Unmarshal. (#425)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusNoContent) // 204, empty body
	})

	_, stderr, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "delete", "A-S001", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stderr, "deleted")
}

func TestDelete_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/subscriptions/A-S001/delete", map[string]interface{}{"success": true})

	_, stderr, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "delete", "A-S001", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stderr, "deleted")
}

func TestDelete_RequiresConfirm(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "delete", "A-S001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}
