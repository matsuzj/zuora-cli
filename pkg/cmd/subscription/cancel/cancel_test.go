package cancel

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCancel(f) }

func TestCancel_WithPolicy(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Contains(t, r.URL.Path, "/cancel")
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "EndOfCurrentTerm", body["cancellationPolicy"])
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "subscriptionId": "sub-1"})
	})

	stdout, stderr, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "cancel", "A-S001", "--policy", "EndOfCurrentTerm", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	assert.Contains(t, stderr, "cancelled")
}

func TestCancel_WithBody(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/subscriptions/A-S001/cancel", map[string]interface{}{"success": true})

	_, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "cancel", "A-S001", "--body", `{"cancellationPolicy":"EndOfCurrentTerm"}`, "--confirm")
	require.NoError(t, err)
}

func TestCancel_RequiresPolicyOrBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "cancel", "A-S001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one of the flags in the group [body policy] is required")
}

func TestCancel_SpecificDateRequiresEffectiveDate(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "cancel", "A-S001", "--policy", "SpecificDate", "--confirm")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--effective-date is required")
}

// TestSubscriptionCancel_ExplicitEmptyPolicyAndBodyRejected pins the P5-2
// edge case: --policy "" satisfies cobra's group check (the flag WAS
// provided) but the disjunction is enforced on the values too.
func TestSubscriptionCancel_ExplicitEmptyPolicyAndBodyRejected(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil,
		"subscription", "cancel", "A-S001", "--policy", "", "--confirm")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one of the flags in the group [body policy] is required")
}
