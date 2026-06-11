package get

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdGet(f) }

func TestRatePlanGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/rateplans/402880e123", map[string]interface{}{
		"id":           "402880e123",
		"ratePlanName": "Monthly Plan",
		"productId":    "prod-001",
		"productName":  "My Product",
		"productSku":   "SKU-1",
		// Real subscription-rate-plan response keys (GET /v1/rateplans/{id}).
		"productRatePlanId":   "PRP-001",
		"subscriptionId":      "sub-001",
		"subscriptionVersion": 99,
	})

	stdout, _, err := cmdtest.Run(t, "rateplan", newCmd, handler, "rateplan", "get", "402880e123")
	require.NoError(t, err)
	assert.Contains(t, stdout, "402880e123")
	assert.Contains(t, stdout, "Monthly Plan")
	assert.Contains(t, stdout, "My Product")
	// Guard every renamed/new key: each distinctive value only renders if the
	// command reads the correct subscription-rate-plan key.
	assert.Contains(t, stdout, "SKU-1")   // productSku
	assert.Contains(t, stdout, "PRP-001") // productRatePlanId
	assert.Contains(t, stdout, "sub-001") // subscriptionId
	assert.Contains(t, stdout, "99")      // subscriptionVersion
}

func TestRatePlanGet_PathEscape(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/rateplans/a%2Fb", r.URL.RawPath)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": "a/b"})
	})

	_, _, err := cmdtest.Run(t, "rateplan", newCmd, handler, "rateplan", "get", "a/b")
	require.NoError(t, err)
}

func TestRatePlanGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "rateplan", newCmd, nil, "rateplan", "get")
	assert.Error(t, err)
}
