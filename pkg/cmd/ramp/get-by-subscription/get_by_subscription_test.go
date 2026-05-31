package getbysubscription

import (
	"encoding/json"
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
	ramp := &cobra.Command{Use: "ramp"}
	ramp.AddCommand(NewCmdGetBySubscription(f))
	root.AddCommand(ramp)
	return root
}

func TestRampGetBySubscription_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/subscriptions/A-S00000001/ramps", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"ramps": []map[string]interface{}{
				{
					"rampNumber":         "R-00000001",
					"name":               "Test Ramp",
					"subscriptionNumber": "A-S00000001",
				},
			},
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"ramp", "get-by-subscription", "A-S00000001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "R-00000001")
	assert.Contains(t, out.String(), "Test Ramp")
}

func TestRampGetBySubscription_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"ramp", "get-by-subscription"})
	err := root.Execute()

	assert.Error(t, err)
}

func TestRampGetBySubscription_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{
				{"code": 50000000, "message": "Subscription not found"},
			},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"ramp", "get-by-subscription", "A-INVALID"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Subscription not found")
}
