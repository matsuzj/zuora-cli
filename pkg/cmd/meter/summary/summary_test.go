package summary

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdSummary(f) }

func TestMeterSummary_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/meters/meter123/summary", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var reqBody map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &reqBody))
		assert.Equal(t, "NORMAL", reqBody["runType"])

		// Doc-verified mediation envelope (#486): {success, data:{requestId,
		// requestTime, query:{runType}, output:[…]}} — the old flat
		// meterId/runType fixture encoded keys the API never returns.
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"requestId":   "req-0042",
				"requestTime": "2026-07-05T10:00:00Z",
				"query":       map[string]interface{}{"runType": "NORMAL"},
				"output": []interface{}{
					map[string]interface{}{"dimensions": map[string]interface{}{"accountId": "a1"}, "output": 7, "totalErrorCount": 0},
				},
			},
		})
	})

	stdout, _, err := cmdtest.Run(t, "meter", newCmd, handler, "meter", "summary", "meter123", "--run-type", "NORMAL")

	require.NoError(t, err)
	assert.Regexp(t, `(?m)^Request ID:\s+req-0042$`, stdout)
	assert.Regexp(t, `(?m)^Request Time:\s+2026-07-05T10:00:00Z$`, stdout)
	assert.Regexp(t, `(?m)^Run Type:\s+NORMAL$`, stdout)
	assert.Regexp(t, `(?m)^Output Groups:\s+1$`, stdout)
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
}

func TestMeterSummary_WithBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var reqBody map[string]interface{}
		require.NoError(t, json.Unmarshal(body, &reqBody))
		assert.Equal(t, "FULL", reqBody["runType"])
		assert.Equal(t, "2026-01-01", reqBody["startDate"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	})

	_, _, err := cmdtest.Run(t, "meter", newCmd, handler, "meter", "summary", "meter123", "--run-type", "FULL", "--body", `{"startDate":"2026-01-01"}`)

	require.NoError(t, err)
}

func TestMeterSummary_RequiresRunType(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil, "meter", "summary", "meter123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "run-type" not set`)
}

func TestMeterSummary_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil, "meter", "summary", "--run-type", "FULL")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
