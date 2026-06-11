package signup

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdSignup(f) }

func TestSignup_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/sign-up", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "accountInfo")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":            true,
			"accountId":          "a-new",
			"accountNumber":      "A00099",
			"subscriptionId":     "s-new",
			"subscriptionNumber": "A-S001",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "", newCmd, handler, "signup", "--body", `{"accountInfo":{},"subscriptionInfo":{}}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "A00099")
	assert.Contains(t, stdout, "A-S001")
	assert.Contains(t, stderr, "Sign-up complete. Account A00099 created.")
}

func TestSignup_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newCmd, nil, "signup")
	assert.Error(t, err)
}
