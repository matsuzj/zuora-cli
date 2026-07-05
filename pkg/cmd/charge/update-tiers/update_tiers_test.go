package updatetiers

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdUpdateTiers(f) }

func TestChargeUpdateTiers_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/commerce/tiers", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		// The --body payload must reach the server intact (#484): the handler
		// previously ignored r.Body.
		body, rerr := io.ReadAll(r.Body)
		if assert.NoError(t, rerr) {
			assert.JSONEq(t, `{"charge_id":"chg-001","tiers":[]}`, string(body))
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "charge", newCmd, handler, "charge", "update-tiers", "--body", `{"charge_id":"chg-001","tiers":[]}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "success")
	// The response body is emitted verbatim (#483): assert the whole envelope
	// structurally, not just the substring "success".
	assert.JSONEq(t, `{"success":true}`, stdout)
	assert.Contains(t, stderr, "Tiers updated.")
}

func TestChargeUpdateTiers_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "charge", newCmd, nil, "charge", "update-tiers")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
