package delete

import (
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
	acct.AddCommand(NewCmdDelete(f))
	root.AddCommand(acct)
	return root
}

func TestAccountDelete_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/accounts/A001", r.URL.Path)
		w.WriteHeader(204)
	}))

	ios, _, _, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "delete", "A001", "--confirm"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, errOut.String(), "Account A001 deleted.")
}

func TestAccountDelete_JSON(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "delete", "A001", "--confirm", "--json"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), `"success": true`)
}

func TestAccountDelete_RequiresConfirm(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "delete", "A001"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}
