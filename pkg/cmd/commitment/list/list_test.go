package list

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdList(f) }

func TestCommitmentList_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/commitments", r.URL.Path)
		assert.Equal(t, "A00000001", r.URL.Query().Get("accountNumber"))
		// Doc-verified envelope (#453): {total, page, page_size, commitments}
		// with flat items — the old fixture's success flag and commitmentKey
		// item field do not exist in the documented 200 schema.
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": 1, "page": 1, "page_size": 20,
			"commitments": []map[string]interface{}{{
				"id":               "cmt-id-1",
				"commitmentNumber": "CMT-00042",
				"name":             "Fixture Commitment",
				"type":             "MinCommitment",
				"status":           "Active",
				"accountNumber":    "A00000001",
				"startDate":        "2026-01-01",
				"endDate":          "2026-12-31",
				"totalAmount":      1200.5,
				"currency":         "JPY",
			}},
		})
	})

	stdout, _, err := cmdtest.Run(t, "commitment", newCmd, handler, "commitment", "list", "--account-number", "A00000001")

	require.NoError(t, err)
	// Table cells for every declared column (#453).
	for _, cell := range []string{"CMT-00042", "Fixture Commitment", "MinCommitment", "Active", "2026-01-01", "2026-12-31", "1200.50", "JPY"} {
		assert.Contains(t, stdout, cell)
	}
}

// TestCommitmentList_TypeFilter pins the --type query parameter wiring and
// the zero-row empty state.
func TestCommitmentList_TypeFilter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "MaxCommitment", r.URL.Query().Get("type"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"commitments": []interface{}{}})
	})
	_, stderr, err := cmdtest.Run(t, "commitment", newCmd, handler, "commitment", "list", "--account-number", "A1", "--type", "MaxCommitment")
	require.NoError(t, err)
	assert.Contains(t, stderr, "No results found.")
}

func TestCommitmentList_RequiresAccount(t *testing.T) {
	_, _, err := cmdtest.Run(t, "commitment", newCmd, nil, "commitment", "list")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--account-number is required")
}

// TestCommitmentList_AccountAliasRemoved pins that the deprecated --account
// alias is gone (v0.7.0): only --account-number is accepted.
func TestCommitmentList_AccountAliasRemoved(t *testing.T) {
	_, _, err := cmdtest.Run(t, "commitment", newCmd, nil, "commitment", "list", "--account", "A00000001", "--json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown flag: --account")
}
