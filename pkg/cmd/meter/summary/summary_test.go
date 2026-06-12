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
		assert.Equal(t, "FULL", reqBody["runType"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"meterId": "meter123",
			"runType": "FULL",
		})
	})

	stdout, _, err := cmdtest.Run(t, "meter", newCmd, handler, "meter", "summary", "meter123", "--run-type", "FULL")

	require.NoError(t, err)
	assert.Contains(t, stdout, "meter123")
	assert.Contains(t, stdout, "FULL")
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
}
