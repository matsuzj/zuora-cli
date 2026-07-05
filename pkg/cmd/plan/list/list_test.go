package list

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdList(f) }

func TestPlanList_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/commerce/plans/list", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		// The --body payload must reach the server intact (#484): the handler
		// previously ignored r.Body, so a command that dropped or mangled the
		// body would still pass.
		body, err := io.ReadAll(r.Body)
		if assert.NoError(t, err) {
			assert.JSONEq(t, `{"filters":[{"field":"state","operator":"EQ","value":"active"}]}`, string(body))
		}
		// Doc-verified envelope (#453): {"values":[...]} with no success flag
		// at 200 — the old {"success","data"} fixture shape does not exist in
		// the documented schema.
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"values": []interface{}{
				map[string]interface{}{
					"id":                    "prp-4f7d2a83",
					"name":                  "Plan List Fixture #483",
					"productRatePlanNumber": "PRP-00000172",
					"productId":             "prod-9b1c",
					"state":                 "active",
					"startDate":             "2026-01-01",
					"endDate":               "2031-12-31",
				},
			},
		})
	})

	stdout, _, err := cmdtest.Run(t, "plan", newCmd, handler, "plan", "list", "--body", `{"filters":[{"field":"state","operator":"EQ","value":"active"}]}`)
	require.NoError(t, err)
	// Table cells for every declared column (#453/#483).
	for _, cell := range []string{"Plan List Fixture #483", "PRP-00000172", "prod-9b1c", "active", "2026-01-01", "2031-12-31", "prp-4f7d2a83"} {
		assert.Contains(t, stdout, cell)
	}
}

// TestPlanList_EmptyValues pins the zero-row table empty state (#453): the
// human table prints "No results found." on stderr with stdout empty, and
// --json still passes the raw envelope through.
func TestPlanList_EmptyValues(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "/commerce/plans/list", map[string]interface{}{
		"values": []interface{}{},
	})

	stdout, stderr, err := cmdtest.Run(t, "plan", newCmd, handler, "plan", "list", "--body", `{"filters":[]}`)
	require.NoError(t, err)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "No results found.")
}

// TestPlanList_JSONPassthrough pins that --json emits the raw envelope.
func TestPlanList_JSONPassthrough(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "/commerce/plans/list", map[string]interface{}{
		"values": []interface{}{},
	})

	stdout, _, err := cmdtest.Run(t, "plan", newCmd, handler, "plan", "list", "--body", `{"filters":[]}`, "--json")
	require.NoError(t, err)
	assert.JSONEq(t, `{"values":[]}`, stdout)
}

func TestPlanList_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "plan", newCmd, nil, "plan", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
