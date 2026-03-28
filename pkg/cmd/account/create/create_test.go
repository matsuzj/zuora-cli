package create

import (
	"encoding/json"
	"github.com/matsuzj/zuora-cli/internal/testutil"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().String("jq", "", "")
	root.PersistentFlags().String("template", "", "")
	acct := &cobra.Command{Use: "account"}
	acct.AddCommand(NewCmdCreate(f))
	root.AddCommand(acct)
	return root
}

func TestAccountCreate_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/accounts", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       true,
			"accountId":     "id-123",
			"accountNumber": "A00099",
		})
	}))

	ios, _, out, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "create", "--body", `{"name":"Test"}`})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "A00099")
	assert.Contains(t, errOut.String(), "Account A00099 created.")
}

func TestAccountCreate_SuccessFalse(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{
				{"code": 53100020, "message": "Missing required field"},
			},
		})
	}))

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "create", "--body", `{"name":"Bad"}`})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Missing required field")
}

func TestAccountCreate_RequiresBody(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "create"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
