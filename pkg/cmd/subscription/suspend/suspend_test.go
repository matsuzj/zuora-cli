package suspend

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdSuspend(f) }

func TestSuspend_WithPolicy(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "FixedPeriodsFromToday", body["suspendPolicy"])
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	})

	_, stderr, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "suspend", "A-S001", "--policy", "FixedPeriodsFromToday", "--periods", "3", "--periods-type", "Month")
	require.NoError(t, err)
	assert.Contains(t, stderr, "suspended")
}

func TestSuspend_RequiresPolicyOrBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "suspend", "A-S001")
	assert.Error(t, err)
}

func TestSuspend_SpecificDateRequiresSuspendDate(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "suspend", "A-S001", "--policy", "SpecificDate")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--suspend-date is required")
}

func TestSuspend_FixedPeriodsRequiresPeriodsAndType(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "suspend", "A-S001", "--policy", "FixedPeriodsFromToday", "--periods", "3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--periods and --periods-type are required")
}
