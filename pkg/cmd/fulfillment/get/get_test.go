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
	fulfillment := &cobra.Command{Use: "fulfillment"}
	fulfillment.AddCommand(NewCmdGet(f))
	root.AddCommand(fulfillment)
	return root
}

func TestFulfillmentGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/fulfillments/F-00000001", r.URL.Path)
		w.WriteHeader(200)
		// Real shape: nested under a "fulfillment" object; the number is
		// "fulfillmentNumber" (no top-level "key") and the date is "fulfillmentDate".
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"fulfillment": map[string]interface{}{
				"id":                "8aca-ful-id",
				"fulfillmentNumber": "F-00000001",
				"state":             "Executed",
				"orderLineItemId":   "OLI-001",
				"quantity":          5,
				"fulfillmentDate":   "2026-05-30",
			},
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"fulfillment", "get", "F-00000001"})
	err := root.Execute()

	require.NoError(t, err)
	outStr := out.String()
	assert.Contains(t, outStr, "F-00000001") // fulfillmentNumber (was read from the absent "key")
	assert.Contains(t, outStr, "Executed")   // nested state
	assert.Contains(t, outStr, "OLI-001")    // nested orderLineItemId
	assert.Contains(t, outStr, "2026-05-30") // fulfillmentDate (was read from the absent "date")
}

func TestFulfillmentGet_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"fulfillment", "get"})
	err := root.Execute()

	assert.Error(t, err)
}

func TestFulfillmentGet_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{
				{"code": 50000040, "message": "Fulfillment not found"},
			},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"fulfillment", "get", "F-99999999"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Fulfillment not found")
}
