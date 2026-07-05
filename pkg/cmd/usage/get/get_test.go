package get

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdGet(f) }

func TestUsageGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/object/usage/2c92a0f96bd", map[string]interface{}{
		"Id":             "2c92a0f96bd",
		"AccountId":      "acc123",
		"Quantity":       10,
		"StartDateTime":  "2026-01-01",
		"UOM":            "Each",
		"RbeStatus":      "Processed", // the Usage CRUD object's status field is "RbeStatus"; "Status" never existed
		"EndDateTime":    "2034-03-02T09:00:00Z",
		"SubscriptionId": "sub-usage-77",
		"ChargeId":       "chg-usage-88",
		"CreatedDate":    "2034-01-11T01:02:03Z",
		"UpdatedDate":    "2034-01-12T04:05:06Z",
	})

	stdout, _, err := cmdtest.Run(t, "usage", newCmd, handler, "usage", "get", "2c92a0f96bd")
	require.NoError(t, err)
	// Label-bound (F-08): values under their own labels.
	assert.Regexp(t, `(?m)^ID:\s+2c92a0f96bd$`, stdout)
	assert.Regexp(t, `(?m)^UOM:\s+Each$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Processed$`, stdout) // sourced from the real "RbeStatus" key
	// Fixture-masking backfill (#482): pin every prod-read key under its label.
	assert.Regexp(t, `(?m)^End Date:\s+2034-03-02T09:00:00Z$`, stdout)
	assert.Regexp(t, `(?m)^Subscription ID:\s+sub-usage-77$`, stdout)
	assert.Regexp(t, `(?m)^Charge ID:\s+chg-usage-88$`, stdout)
	assert.Regexp(t, `(?m)^Created Date:\s+2034-01-11T01:02:03Z$`, stdout)
	assert.Regexp(t, `(?m)^Updated Date:\s+2034-01-12T04:05:06Z$`, stdout)
}

func TestUsageGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "usage", newCmd, nil, "usage", "get")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
