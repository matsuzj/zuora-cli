package resume

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdResume(f) }

func TestResume_WithPolicy(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "SpecificDate", body["resumePolicy"])
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	})

	_, stderr, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "resume", "A-S001", "--policy", "SpecificDate", "--resume-date", "2026-05-01")
	require.NoError(t, err)
	assert.Contains(t, stderr, "resumed")
}

func TestResume_RequiresPolicyOrBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "resume", "A-S001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one of the flags in the group [body policy] is required")
}

func TestResume_SpecificDateRequiresResumeDate(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "resume", "A-S001", "--policy", "SpecificDate")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--resume-date is required")
}

func TestResume_FixedPeriodsRequiresPeriodsAndType(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "resume", "A-S001", "--policy", "FixedPeriodsFromSuspendDate", "--periods", "1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--periods and --periods-type are required")
}
