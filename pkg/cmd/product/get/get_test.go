package get

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
	product := &cobra.Command{Use: "product"}
	product.AddCommand(NewCmdGet(f))
	root.AddCommand(product)
	return root
}

func TestProductGet_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/commerce/products/PROD-001", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "prod-001",
			"name":        "My Product",
			"sku":         "SKU-001",
			"description": "A test product",
		})
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"product", "get", "PROD-001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "My Product")
	assert.Contains(t, out.String(), "prod-001")
}

func TestProductGet_PathEscape(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/commerce/products/a%2Fb", r.URL.RawPath)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": "a/b"})
	}))

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"product", "get", "a/b"})
	err := root.Execute()

	require.NoError(t, err)
}

func TestProductGet_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"product", "get"})
	err := root.Execute()

	assert.Error(t, err)
}
