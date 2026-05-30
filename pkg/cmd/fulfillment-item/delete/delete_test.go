package delete

import (
	"net/http"
	"net/http/httptest"
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
	fi := &cobra.Command{Use: "fulfillment-item"}
	fi.AddCommand(NewCmdDelete(f))
	root.AddCommand(fi)
	return root
}

func TestFulfillmentItemDelete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/fulfillment-items/item-001", r.URL.Path)
		w.WriteHeader(204)
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"fulfillment-item", "delete", "item-001", "--confirm"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, errOut.String(), "item-001")
	assert.Contains(t, errOut.String(), "deleted")
}

func TestFulfillmentItemDelete_RequiresConfirm(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(204)
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"fulfillment-item", "delete", "item-001"})
	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
	assert.False(t, called, "no HTTP call should be made when --confirm is omitted")
}

func TestFulfillmentItemDelete_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"fulfillment-item", "delete", "--confirm"})
	err := root.Execute()

	assert.Error(t, err)
}
