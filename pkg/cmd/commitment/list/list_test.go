package list

import (
	"encoding/json"
	"fmt"
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
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"commitments": []map[string]interface{}{
				{"commitmentKey": "CMT-001", "name": "Test Commitment"},
			},
		})
	})

	stdout, _, err := cmdtest.Run(t, "commitment", newCmd, handler, "commitment", "list", "--account-number", "A00000001")

	require.NoError(t, err)
	assert.Contains(t, stdout, "CMT-001")
}

func TestCommitmentList_RequiresAccount(t *testing.T) {
	_, _, err := cmdtest.Run(t, "commitment", newCmd, nil, "commitment", "list")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--account-number is required")
}

// TestCommitmentList_DeprecatedAccountAliasStillWorks pins the P5-1
// deprecation contract for the handwritten (cmdutil.AddAccountNumberFlag)
// path: --account keeps feeding accountNumber through v0.5.x.
func TestCommitmentList_DeprecatedAccountAliasStillWorks(t *testing.T) {
	var gotAccount string
	handler := func(w http.ResponseWriter, r *http.Request) {
		gotAccount = r.URL.Query().Get("accountNumber")
		fmt.Fprint(w, `{"commitments": []}`)
	}

	_, _, err := cmdtest.Run(t, "commitment", newCmd, handler, "commitment", "list", "--account", "A00000001", "--json")
	require.NoError(t, err)
	assert.Equal(t, "A00000001", gotAccount)
}
