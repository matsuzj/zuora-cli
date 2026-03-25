package metrics

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
	sub := &cobra.Command{Use: "subscription"}
	sub.AddCommand(NewCmdMetrics(f))
	root.AddCommand(sub)
	return root
}

func TestSubscriptionMetrics_Table(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/subscriptions/subscription-metrics", r.URL.Path)
		assert.Equal(t, "A-S001,A-S002", r.URL.Query().Get("subscriptionNumbers"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"subscriptionMetrics": []map[string]interface{}{
				{"subscriptionNumber": "A-S001", "mrr": 100.0, "tcv": 1200.0, "tcb": 600.0, "currency": "USD"},
				{"subscriptionNumber": "A-S002", "mrr": 50.0, "tcv": 600.0, "tcb": 300.0, "currency": "USD"},
			},
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "metrics", "--subscription-numbers", "A-S001,A-S002"})
	err := root.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "A-S001")
	assert.Contains(t, output, "100.00")
	assert.Contains(t, output, "A-S002")
}

func TestSubscriptionMetrics_RequiresFlag(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "metrics"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subscription-numbers")
}
