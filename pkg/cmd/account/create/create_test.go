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

func TestAccountCreate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/accounts", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       true,
			"accountId":     "id-123",
			"accountNumber": "A00099",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "account", newCmd, handler, "account", "create", "--body", `{"name":"Test"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "A00099")
	assert.Contains(t, stderr, "Account A00099 created.")
}

func TestAccountCreate_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Missing required field")

	_, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "create", "--body", `{"name":"Bad"}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Missing required field")
}

func TestAccountCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "account", newCmd, nil, "account", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
