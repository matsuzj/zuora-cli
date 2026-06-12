package update

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdUpdate(f) }

func TestChargeUpdate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/commerce/charges", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "chg-001",
			"name": "Updated Charge",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "charge", newCmd, handler, "charge", "update", "--body", `{"id":"chg-001","name":"Updated Charge"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "chg-001")
	assert.Contains(t, stdout, "Updated Charge")
	assert.Contains(t, stderr, "Charge updated.")
}

func TestChargeUpdate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "charge", newCmd, nil, "charge", "update")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
