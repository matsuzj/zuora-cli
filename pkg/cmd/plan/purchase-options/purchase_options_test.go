package purchaseoptions

import (
	"encoding/json"
	httptest "github.com/matsuzj/zuora-cli/internal/testutil/httpmock"
	"io"
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
	plan := &cobra.Command{Use: "plan"}
	plan.AddCommand(NewCmdPurchaseOptions(f))
	root.AddCommand(plan)
	return root
}

func TestPlanPurchaseOptions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/commerce/purchase-options/list", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var payload map[string]interface{}
		err = json.Unmarshal(body, &payload)
		require.NoError(t, err)

		filters, ok := payload["filters"].([]interface{})
		require.True(t, ok, "filters should be an array")
		require.Len(t, filters, 1)

		filter := filters[0].(map[string]interface{})
		assert.Equal(t, "prp_id", filter["field"])
		assert.Equal(t, "=", filter["operator"])

		value := filter["value"].(map[string]interface{})
		assert.Equal(t, "plan-123", value["string_value"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    []interface{}{},
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"plan", "purchase-options", "--plan", "plan-123"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "success")
}

func TestPlanPurchaseOptions_RequiresPlan(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"plan", "purchase-options"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--plan is required")
}
