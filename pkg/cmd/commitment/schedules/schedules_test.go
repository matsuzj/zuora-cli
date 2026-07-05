package schedules

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdSchedules(f) }

func TestCommitmentSchedules_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/commitments/CMT-00000001/schedules", map[string]interface{}{
		"success": true,
		"schedules": []map[string]interface{}{
			{
				"scheduleNumber": "SCH-00000001",
				"startDate":      "2026-01-01",
				"endDate":        "2026-12-31",
				"amount":         1000,
			},
		},
	})

	stdout, _, err := cmdtest.Run(t, "commitment", newCmd, handler, "commitment", "schedules", "CMT-00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "SCH-00000001")
	assert.Contains(t, stdout, "schedules")
}

func TestCommitmentSchedules_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "commitment", newCmd, nil, "commitment", "schedules")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
