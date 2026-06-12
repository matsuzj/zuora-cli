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
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":         true,
			"subscriptionKey": "S-001",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "omnichannel", newCmd, handler, "omnichannel", "create", "--body", `{"subscriptionKey":"S-001"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "S-001")
	assert.Contains(t, stderr, "Omni-channel subscription S-001 created.")
}

func TestOmnichannelCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "omnichannel", newCmd, nil, "omnichannel", "create")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
