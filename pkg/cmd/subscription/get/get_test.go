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

func TestSubscriptionGet_Detail(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/subscriptions/A-S001", map[string]interface{}{
		"id": "sub-1", "subscriptionNumber": "A-S001", "name": "Gold Plan",
		"status": "Active", "accountId": "acct-1", "termType": "TERMED",
		"termStartDate": "2025-01-01", "termEndDate": "2026-01-01",
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "get", "A-S001")
	require.NoError(t, err)
	// Label-bound (F-08): values under their own labels.
	assert.Regexp(t, `(?m)^Name:\s+Gold Plan$`, stdout)
	assert.Regexp(t, `(?m)^Subscription Number:\s+A-S001$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Active$`, stdout)
}

func TestSubscriptionGet_SuccessFalse(t *testing.T) {
	// HTTP 200 with a success:false envelope must be treated as an error
	// (the success-flag check is on by default in the API client).
	handler := cmdtest.Reasons(t, 50000040, "Subscription not found")

	_, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "get", "A-S001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Subscription not found")
}

func TestSubscriptionGet_JSON(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"id": "sub-1", "name": "Gold Plan",
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "get", "A-S001", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"name"`)
}
