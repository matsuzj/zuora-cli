package periods

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdPeriods(f) }

func TestCommitmentPeriods_ByCommitment_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/commitments/periods", r.URL.Path)
		assert.Equal(t, "CMT-00000001", r.URL.Query().Get("commitmentKey"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"periods": []map[string]interface{}{
				{
					"periodId":      "PRD-00000001",
					"commitmentKey": "CMT-00000001",
					"startDate":     "2026-01-01",
					"endDate":       "2026-12-31",
				},
			},
		})
	})

	stdout, _, err := cmdtest.Run(t, "commitment", newCmd, handler, "commitment", "periods", "--commitment", "CMT-00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "PRD-00000001")
	assert.Contains(t, stdout, "CMT-00000001")
}

func TestCommitmentPeriods_ByAccount_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/commitments/periods", r.URL.Path)
		assert.Equal(t, "A00000001", r.URL.Query().Get("accountNumber"))
		assert.Equal(t, "2026-01-01", r.URL.Query().Get("startDate"))
		assert.Equal(t, "2026-12-31", r.URL.Query().Get("endDate"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"periods": []map[string]interface{}{
				{
					"periodId":      "PRD-00000002",
					"accountNumber": "A00000001",
				},
			},
		})
	})

	stdout, _, err := cmdtest.Run(t, "commitment", newCmd, handler, "commitment", "periods", "--account", "A00000001", "--start-date", "2026-01-01", "--end-date", "2026-12-31")
	require.NoError(t, err)
	assert.Contains(t, stdout, "PRD-00000002")
	assert.Contains(t, stdout, "A00000001")
}

func TestCommitmentPeriods_RequiresCommitmentOrAccount(t *testing.T) {
	_, _, err := cmdtest.Run(t, "commitment", newCmd, nil, "commitment", "periods")
	assert.Error(t, err)
}

func TestCommitmentPeriods_MutuallyExclusive(t *testing.T) {
	_, _, err := cmdtest.Run(t, "commitment", newCmd, nil, "commitment", "periods", "--commitment", "CMT-00000001", "--account", "A00000001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestCommitmentPeriods_AccountRequiresDates(t *testing.T) {
	_, _, err := cmdtest.Run(t, "commitment", newCmd, nil, "commitment", "periods", "--account", "A00000001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start-date")
}
