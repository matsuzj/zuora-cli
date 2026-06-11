package metrics

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdMetrics(f) }

func TestSubscriptionMetrics_Table(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/subscriptions/subscription-metrics", r.URL.Path)
		assert.Equal(t, []string{"A-S001", "A-S002"}, r.URL.Query()["subscriptionNumbers[]"])
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"subscriptionMetrics": []map[string]interface{}{
				{"subscriptionNumber": "A-S001", "contractedMrr": 100.0, "asOfDayGrossMrr": 100.0, "asOfDayNetMrr": 100.0, "totalContractedValue": 1200.0, "netTotalContractedValue": 1200.0},
				{"subscriptionNumber": "A-S002", "contractedMrr": 50.0, "asOfDayGrossMrr": 50.0, "asOfDayNetMrr": 50.0, "totalContractedValue": 600.0, "netTotalContractedValue": 600.0},
			},
		})
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "metrics", "--subscription-numbers", "A-S001,A-S002")
	require.NoError(t, err)
	assert.Contains(t, stdout, "A-S001")
	assert.Contains(t, stdout, "100.00")
	assert.Contains(t, stdout, "A-S002")
}

func TestSubscriptionMetrics_RequiresFlag(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "metrics")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subscription-numbers")
}

func TestSubscriptionMetrics_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Subscription number is invalid")

	_, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "metrics", "--subscription-numbers", "INVALID")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Subscription number is invalid")
}
