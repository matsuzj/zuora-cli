package rollover

import (
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdRollover(f) }

func TestPrepaidRollover_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/ppdd/rollover", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		cmdtest.OK(t, "", "", map[string]interface{}{
			"success": true,
		})(w, r)
	})

	_, stderr, err := cmdtest.Run(t, "prepaid", newCmd, handler, "prepaid", "rollover", "--body", `{"subscriptionNumber":"A-S001"}`)
	require.NoError(t, err)
	assert.Contains(t, stderr, "Prepaid rollover completed.")
}

func TestPrepaidRollover_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "prepaid", newCmd, nil, "prepaid", "rollover")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestPrepaidRollover_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000000, "Rollover failed")

	_, _, err := cmdtest.Run(t, "prepaid", newCmd, handler, "prepaid", "rollover", "--body", `{"bad":"data"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Rollover failed")
}
