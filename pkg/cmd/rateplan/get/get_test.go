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
	rateplan := &cobra.Command{Use: "rateplan"}
	rateplan.AddCommand(NewCmdGet(f))
	root.AddCommand(rateplan)
	return root
}

func TestRatePlanGet_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/rateplans/402880e123", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":                    "402880e123",
			"name":                  "Monthly Plan",
			"productId":             "prod-001",
			"productName":           "My Product",
			"status":                "Active",
			"productRatePlanNumber": "PRP-001",
		})
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"rateplan", "get", "402880e123"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "402880e123")
	assert.Contains(t, out.String(), "Monthly Plan")
	assert.Contains(t, out.String(), "My Product")
}

func TestRatePlanGet_PathEscape(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/rateplans/a%2Fb", r.URL.RawPath)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": "a/b"})
	}))

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"rateplan", "get", "a/b"})
	err := root.Execute()

	require.NoError(t, err)
}

func TestRatePlanGet_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"rateplan", "get"})
	err := root.Execute()

	assert.Error(t, err)
}
