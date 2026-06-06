package get

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
	rateplan := &cobra.Command{Use: "rateplan"}
	rateplan.AddCommand(NewCmdGet(f))
	root.AddCommand(rateplan)
	return root
}

func TestRatePlanGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/rateplans/402880e123", r.URL.Path)
		w.WriteHeader(200)
		// Real subscription-rate-plan response keys (GET /v1/rateplans/{id}).
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":                  "402880e123",
			"ratePlanName":        "Monthly Plan",
			"productId":           "prod-001",
			"productName":         "My Product",
			"productSku":          "SKU-1",
			"productRatePlanId":   "PRP-001",
			"subscriptionId":      "sub-001",
			"subscriptionVersion": 99,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"rateplan", "get", "402880e123"})
	err := root.Execute()

	require.NoError(t, err)
	outStr := out.String()
	assert.Contains(t, outStr, "402880e123")
	assert.Contains(t, outStr, "Monthly Plan")
	assert.Contains(t, outStr, "My Product")
	// Guard every renamed/new key: each distinctive value only renders if the
	// command reads the correct subscription-rate-plan key.
	assert.Contains(t, outStr, "SKU-1")   // productSku
	assert.Contains(t, outStr, "PRP-001") // productRatePlanId
	assert.Contains(t, outStr, "sub-001") // subscriptionId
	assert.Contains(t, outStr, "99")      // subscriptionVersion
}

func TestRatePlanGet_PathEscape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/rateplans/a%2Fb", r.URL.RawPath)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": "a/b"})
	}))
	defer server.Close()

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
