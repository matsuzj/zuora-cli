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
			assert.JSONEq(t, `{"page":0,"page_size":20}`, string(body))
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": []interface{}{
				map[string]interface{}{
					"id":   "prp-4f7d2a83",
					"name": "Plan List Fixture #483",
				},
			},
		})
	})

	stdout, _, err := cmdtest.Run(t, "plan", newCmd, handler, "plan", "list", "--body", `{"page":0,"page_size":20}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "success")
	// Distinctive row VALUES must be rendered (#483): the old empty-data
	// fixture plus a bare Contains("success") passed for ANY non-crash
	// rendering. Commerce fixtures are not live-verifiable (404 on this
	// tenant), so only the values are distinctive; the key shapes are kept.
	assert.Contains(t, stdout, "prp-4f7d2a83")
	assert.Contains(t, stdout, "Plan List Fixture #483")
}

// TestPlanList_EmptyData pins the empty-result rendering of this JSON-only
// command: the full envelope is emitted verbatim (there is no table
// empty-state here), asserted structurally rather than via a substring.
func TestPlanList_EmptyData(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "/commerce/plans/list", map[string]interface{}{
		"success": true,
		"data":    []interface{}{},
	})

	stdout, _, err := cmdtest.Run(t, "plan", newCmd, handler, "plan", "list", "--body", `{"page":0,"page_size":20}`)
	require.NoError(t, err)
	assert.JSONEq(t, `{"success":true,"data":[]}`, stdout)
}

func TestPlanList_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "plan", newCmd, nil, "plan", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
