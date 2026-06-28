package create

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCreate(f) }

func TestFulfillmentCreate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/fulfillments", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		// Real shape: bulk endpoint returns the created object under a
		// "fulfillments" array (keyed by id/fulfillmentNumber), not a flat "key".
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"fulfillments": []map[string]interface{}{
				{"id": "8aca-ful-id", "fulfillmentNumber": "F-00000001"},
			},
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "fulfillment", newCmd, handler, "fulfillment", "create", "--body", `{"orderLineItemId":"OLI-001"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "F-00000001")
	assert.Contains(t, stderr, "Fulfillment F-00000001 created.")
}

func TestFulfillmentCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "fulfillment", newCmd, nil, "fulfillment", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestFulfillmentCreate_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Invalid fulfillment data")

	_, _, err := cmdtest.Run(t, "fulfillment", newCmd, handler, "fulfillment", "create", "--body", `{"bad":"data"}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid fulfillment data")
}

func TestFulfillmentCreate_EmptyArraySuppressesSuccessMsg(t *testing.T) {
	// Bulk endpoint returned no fulfillments: no id/number, so the "created."
	// confirmation must be suppressed.
	handler := cmdtest.OK(t, "POST", "/v1/fulfillments", map[string]interface{}{
		"success":      true,
		"fulfillments": []map[string]interface{}{},
	})

	_, stderr, err := cmdtest.Run(t, "fulfillment", newCmd, handler, "fulfillment", "create", "--body", `{"orderLineItemId":"OLI-001"}`)
	require.NoError(t, err)
	assert.NotContains(t, stderr, "created.", "empty bulk array → no success confirmation")
}

func TestFulfillmentCreate_MultipleRendersFirst(t *testing.T) {
	// Bulk array with >1 fulfillment: only the FIRST is rendered (firstFulfillment).
	handler := cmdtest.OK(t, "POST", "/v1/fulfillments", map[string]interface{}{
		"success": true,
		"fulfillments": []map[string]interface{}{
			{"id": "id-1", "fulfillmentNumber": "F-00000001"},
			{"id": "id-2", "fulfillmentNumber": "F-00000002"},
		},
	})

	stdout, stderr, err := cmdtest.Run(t, "fulfillment", newCmd, handler, "fulfillment", "create", "--body", `{"orderLineItemId":"OLI-001"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "F-00000001")
	assert.NotContains(t, stdout, "F-00000002", "only the first created fulfillment is rendered")
	assert.Contains(t, stderr, "Fulfillment F-00000001 created.")
}
