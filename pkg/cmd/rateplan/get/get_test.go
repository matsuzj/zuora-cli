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
	// Label-bound (F-08): each distinctive value only renders under its OWN label
	// when the command reads the correct subscription-rate-plan key.
	assert.Regexp(t, `(?m)^ID:\s+402880e123$`, stdout)
	assert.Regexp(t, `(?m)^Rate Plan Name:\s+Monthly Plan$`, stdout)
	assert.Regexp(t, `(?m)^Product Name:\s+My Product$`, stdout)
	assert.Regexp(t, `(?m)^Product SKU:\s+SKU-1$`, stdout)
	assert.Regexp(t, `(?m)^Product Rate Plan ID:\s+PRP-001$`, stdout)
	assert.Regexp(t, `(?m)^Subscription ID:\s+sub-001$`, stdout)
	assert.Regexp(t, `(?m)^Subscription Version:\s+99$`, stdout)
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
