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

func TestOmnichannelCreate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/omni-channel-subscriptions", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		// Doc-verified POST response (#414): subscriptionId/subscriptionNumber/
		// accountNumber — the old subscriptionKey key does not exist, so the
		// success message never fired.
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":            true,
			"subscriptionId":     "omni-id-1",
			"subscriptionNumber": "OCS-00042",
			"accountNumber":      "A00001234",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "omnichannel", newCmd, handler, "omnichannel", "create", "--body", `{"externalSubscriptionId":"ext-sub-1","externalSourceSystem":"AppleAppStore"}`)
	require.NoError(t, err)
	assert.Regexp(t, `(?m)^Subscription ID:\s+omni-id-1$`, stdout)
	assert.Regexp(t, `(?m)^Subscription Number:\s+OCS-00042$`, stdout)
	assert.Regexp(t, `(?m)^Account Number:\s+A00001234$`, stdout)
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
	assert.Contains(t, stderr, "Omni-channel subscription OCS-00042 created.")
}

func TestOmnichannelCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "omnichannel", newCmd, nil, "omnichannel", "create")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
