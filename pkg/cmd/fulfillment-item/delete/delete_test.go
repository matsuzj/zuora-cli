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

func TestFulfillmentItemDelete_Success(t *testing.T) {
	handler := cmdtest.Status(t, "DELETE", "/v1/fulfillment-items/item-001", 204, nil)

	_, stderr, err := cmdtest.Run(t, "fulfillment-item", newCmd, handler, "fulfillment-item", "delete", "item-001", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stderr, "item-001")
	assert.Contains(t, stderr, "deleted")
}

func TestFulfillmentItemDelete_RequiresConfirm(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(204)
	})

	_, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, handler, "fulfillment-item", "delete", "item-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
	assert.False(t, called, "no HTTP call should be made when --confirm is omitted")
}

func TestFulfillmentItemDelete_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, nil, "fulfillment-item", "delete", "--confirm")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestFulfillmentItemDelete_JSON(t *testing.T) {
	handler := cmdtest.Status(t, "", "", 204, nil)

	stdout, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, handler, "fulfillment-item", "delete", "item-1", "--confirm", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"success": true`)
}

func TestFulfillmentItemDelete_BodyResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	})

	stdout, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, handler, "fulfillment-item", "delete", "item-1", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	// Label-bound (#483): the delete detail view's only row is Success — a bare
	// Contains "true" would pass on any stray "true" anywhere in the output.
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
}

func TestFulfillmentItemDelete_NonJSONBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	})

	_, stderr, err := cmdtest.Run(t, "fulfillment-item", newCmd, handler, "fulfillment-item", "delete", "item-1", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stderr, "Fulfillment item item-1 deleted.")
}

func TestFulfillmentItemDelete_UnparseableBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[1,2,3]`))
	})

	_, _, err := cmdtest.Run(t, "fulfillment-item", newCmd, handler, "fulfillment-item", "delete", "item-1", "--confirm")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing response")
}
