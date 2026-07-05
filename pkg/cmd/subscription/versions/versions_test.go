package versions

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdVersions(f) }

func TestSubscriptionVersions_Detail(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/subscriptions/A-S001/versions/1", map[string]interface{}{
		"id": "sub-1", "subscriptionNumber": "A-S001", "version": 1,
		"status": "Active", "termType": "TERMED",
		"termStartDate": "2032-02-01", "termEndDate": "2033-02-01",
		"contractEffectiveDate": "2032-01-15",
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "versions", "A-S001", "1")
	require.NoError(t, err)
	assert.Regexp(t, `(?m)^Subscription Number:\s+A-S001$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Active$`, stdout)
	// Fixture-masking backfill (#482): pin every prod-read key under its label.
	assert.Regexp(t, `(?m)^Term Start Date:\s+2032-02-01$`, stdout)
	assert.Regexp(t, `(?m)^Term End Date:\s+2033-02-01$`, stdout)
	assert.Regexp(t, `(?m)^Contract Effective Date:\s+2032-01-15$`, stdout)
	// No top-level "name" in the response (live-verified); the redundant blank
	// Name row was removed. Bites if reintroduced. (#438)
	assert.NotRegexp(t, `(?m)^Name:\s`, stdout)
}

func TestSubscriptionVersions_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "versions", "A-S001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 2 arg(s), received 1")
}

func TestSubscriptionVersions_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Subscription version not found")

	_, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "versions", "bad-key", "99")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Subscription version not found")
}
