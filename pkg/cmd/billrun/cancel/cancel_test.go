package cancel

import (
	"io"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCancel(f) }

func TestBillRunCancel_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/bill-runs/br-001/cancel", map[string]interface{}{
		"id":            "br-001",
		"billRunNumber": "BR-CANCEL-0042",
		"status":        "Cancelled",
		"success":       true,
	})

	stdout, _, err := cmdtest.Run(t, "billrun", newCmd, handler, "billrun", "cancel", "br-001", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Cancelled")
	// Label-bound (F-08) asserts (#482): billRunNumber was never fixtured here,
	// and id/status/success were rendered but unasserted per label — a key typo
	// would render "" while the test stayed green.
	assert.Regexp(t, `(?m)^ID:\s+br-001$`, stdout)
	assert.Regexp(t, `(?m)^Bill Run Number:\s+BR-CANCEL-0042$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Cancelled$`, stdout)
	// success is a JSON boolean read via GetString — renders fmt %v "true".
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
}

func TestBillRunCancel_RequiresConfirm(t *testing.T) {
	cmdtest.RequiresConfirm(t, "billrun", newCmd, "billrun", "cancel", "br-001")
}

func TestBillRunCancel_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "billrun", newCmd, nil, "billrun", "cancel")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

// TestBillRunCancel_SendsEmptyJSONBody pins the 415 fix: Zuora's endpoint binds a Map body parameter
// and rejects requests without a Content-Type, which the client sets only
// when a body is present — the command must send an explicit "{}".
func TestBillRunCancel_SendsEmptyJSONBody(t *testing.T) {
	inner := cmdtest.OK(t, "PUT", "/v1/bill-runs/br-001/cancel", map[string]interface{}{
		"id": "br-001", "status": "Cancelled", "success": true,
	})
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, "{}", string(b))
		inner(w, r)
	}

	_, _, err := cmdtest.Run(t, "billrun", newCmd, handler, "billrun", "cancel", "br-001", "--confirm")
	require.NoError(t, err)
}
