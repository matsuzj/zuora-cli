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

func TestFulfillmentDelete_Success204(t *testing.T) {
	handler := cmdtest.Status(t, "DELETE", "/v1/fulfillments/F-00000001", http.StatusNoContent, nil)

	_, stderr, err := cmdtest.Run(t, "fulfillment", newCmd, handler, "fulfillment", "delete", "F-00000001", "--confirm")

	require.NoError(t, err)
	assert.Contains(t, stderr, "Fulfillment F-00000001 deleted.")
}

func TestFulfillmentDelete_RequiresConfirm(t *testing.T) {
	_, _, err := cmdtest.Run(t, "fulfillment", newCmd, nil, "fulfillment", "delete", "F-00000001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestFulfillmentDelete_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "fulfillment", newCmd, nil, "fulfillment", "delete", "--confirm")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestFulfillmentDelete_JSON(t *testing.T) {
	handler := cmdtest.Status(t, "", "", http.StatusNoContent, nil)

	stdout, _, err := cmdtest.Run(t, "fulfillment", newCmd, handler, "fulfillment", "delete", "F-1", "--confirm", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"success": true`)
}

func TestFulfillmentDelete_BodyResponse(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{"success": true})

	stdout, _, err := cmdtest.Run(t, "fulfillment", newCmd, handler, "fulfillment", "delete", "F-1", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	// Label-bound (#483): the delete detail view's only row is Success — a bare
	// Contains "true" would pass on any stray "true" anywhere in the output.
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
}

func TestFulfillmentDelete_NonJSONBody(t *testing.T) {
	_, stderr, err := cmdtest.Run(t, "fulfillment", newCmd, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}, "fulfillment", "delete", "F-1", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stderr, "Fulfillment F-1 deleted.")
}

func TestFulfillmentDelete_UnparseableBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "fulfillment", newCmd, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[1,2,3]`))
	}, "fulfillment", "delete", "F-1", "--confirm")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing response")
}
