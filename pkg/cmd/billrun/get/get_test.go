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

func TestBillRunGet_Success(t *testing.T) {
	// billCycleDay: numeric per current GetDecimal read — shape unverified live, see #486.
	handler := cmdtest.OK(t, "GET", "/v1/bill-runs/br-001", map[string]interface{}{
		"id":                     "br-001",
		"billRunNumber":          "BR-00000001",
		"name":                   "Fixture bill run #482",
		"status":                 "Completed",
		"invoiceDate":            "2026-05-15",
		"targetDate":             "2026-05-31",
		"autoPost":               false,
		"autoEmail":              true,
		"billCycleDay":           17,
		"scheduledExecutionTime": "2026-05-30T21:30:00Z",
		"createdDate":            "2026-05-01T08:00:00Z",
		"success":                true,
	})

	stdout, _, err := cmdtest.Run(t, "billrun", newCmd, handler, "billrun", "get", "br-001")
	require.NoError(t, err)
	// Label-bound (F-08): values under their own labels.
	assert.Regexp(t, `(?m)^Bill Run Number:\s+BR-00000001$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Completed$`, stdout)
	// autoPost/autoEmail are real JSON booleans (live-verified) — GetBool renders
	// them as true/false. Exercises the boolean fields that the prior fixture
	// never included. (#431)
	assert.Regexp(t, `(?m)^Auto Post:\s+false$`, stdout)
	assert.Regexp(t, `(?m)^Auto Email:\s+true$`, stdout)
	// Backfilled keys (#482): name/invoiceDate/billCycleDay/scheduledExecutionTime/
	// createdDate previously appeared in no fixture — a key typo would render ""
	// while the test stayed green.
	assert.Regexp(t, `(?m)^Name:\s+Fixture bill run #482$`, stdout)
	assert.Regexp(t, `(?m)^Invoice Date:\s+2026-05-15$`, stdout)
	assert.Regexp(t, `(?m)^Bill Cycle Day:\s+17$`, stdout) // GetDecimal: JSON number -> plain decimal
	assert.Regexp(t, `(?m)^Scheduled Execution Time:\s+2026-05-30T21:30:00Z$`, stdout)
	assert.Regexp(t, `(?m)^Created Date:\s+2026-05-01T08:00:00Z$`, stdout)
}

func TestBillRunGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "billrun", newCmd, nil, "billrun", "get")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestBillRunGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Bill run not found")

	_, _, err := cmdtest.Run(t, "billrun", newCmd, handler, "billrun", "get", "bad")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Bill run not found")
}
