package list

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdList(f) }

func TestSubscriptionList_Table(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/subscriptions/accounts/A001", map[string]interface{}{
		"subscriptions": []map[string]interface{}{
			{
				"id": "sub-1", "subscriptionNumber": "A-S001", "name": "Gold Plan",
				"status": "Active", "termType": "TERMED",
				"termStartDate": "2025-01-01", "termEndDate": "2026-01-01",
			},
		},
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "list", "--account", "A001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Gold Plan")
	assert.Contains(t, stdout, "A-S001")
	assert.Contains(t, stdout, "Active")
}

func TestSubscriptionList_JSON(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"subscriptions": []map[string]interface{}{
			{"id": "sub-1", "name": "Gold Plan"},
		},
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "list", "--account", "A001", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"name"`)
}

func TestSubscriptionList_SuccessFalse(t *testing.T) {
	// HTTP 200 with a success:false envelope must be treated as an error
	// (the success-flag check is on by default in the API client).
	handler := cmdtest.Reasons(t, 50000040, "Account not found")

	_, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "list", "--account", "A001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Account not found")
}

func TestSubscriptionList_RequiresAccount(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account")
}
