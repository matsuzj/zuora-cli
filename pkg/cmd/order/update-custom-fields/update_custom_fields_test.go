package updatecustomfields

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdUpdateCustomFields(f) }

func TestOrderUpdateCustomFields_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/orders/O-00000001/customFields", r.URL.Path)

		raw, _ := io.ReadAll(r.Body)
		var sent map[string]interface{}
		require.NoError(t, json.Unmarshal(raw, &sent))
		assert.Equal(t, "value", sent["cf_MyField__c"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "update-custom-fields", "O-00000001", "--body", `{"cf_MyField__c":"value"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	// Label-bound (#483): the command's only detail row is Success, read from
	// the response "success" key — a bare Contains "true" would pass on any
	// stray "true" anywhere in the output.
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
	assert.Contains(t, stderr, "Custom fields updated for order O-00000001.")
}

func TestOrderUpdateCustomFields_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "update-custom-fields", "O-00000001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestOrderUpdateCustomFields_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "update-custom-fields")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
