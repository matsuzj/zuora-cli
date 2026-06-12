package deplete

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdDeplete(f) }

func TestPrepaidDeplete_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/prepaid-balance-funds/deplete", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "USD", body["currency"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	})

	_, stderr, err := cmdtest.Run(t, "prepaid", newCmd, handler, "prepaid", "deplete", "--body", `{"amount":100,"currency":"USD"}`, "--confirm")

	require.NoError(t, err)
	assert.Contains(t, stderr, "Prepaid balance depleted.")
}

func TestPrepaidDeplete_RequiresConfirm(t *testing.T) {
	_, _, err := cmdtest.Run(t, "prepaid", newCmd, nil, "prepaid", "deplete", "--body", `{"amount":100}`)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestPrepaidDeplete_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "prepaid", newCmd, nil, "prepaid", "deplete")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestPrepaidDeplete_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000000, "Deplete failed")

	_, _, err := cmdtest.Run(t, "prepaid", newCmd, handler, "prepaid", "deplete", "--body", `{"bad":"data"}`, "--confirm")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Deplete failed")
}
