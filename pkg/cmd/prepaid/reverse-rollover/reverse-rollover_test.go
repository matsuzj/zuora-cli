package reverserollover

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdReverseRollover(f) }

func TestPrepaidReverseRollover_Success(t *testing.T) {
	var gotBody map[string]interface{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/ppdd/reverse-rollover", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	})

	_, stderr, err := cmdtest.Run(t, "prepaid", newCmd, handler, "prepaid", "reverse-rollover", "--body", `{"subscriptionNumber":"A-S001"}`, "--confirm")

	require.NoError(t, err)
	assert.Equal(t, "A-S001", gotBody["subscriptionNumber"])
	assert.Contains(t, stderr, "Prepaid reverse rollover completed.")
}

func TestPrepaidReverseRollover_RequiresConfirm(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(200)
	})

	_, _, err := cmdtest.Run(t, "prepaid", newCmd, handler, "prepaid", "reverse-rollover", "--body", `{"subscriptionNumber":"A-S001"}`)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
	assert.False(t, called, "no HTTP call should be made without --confirm")
}

func TestPrepaidReverseRollover_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "prepaid", newCmd, nil, "prepaid", "reverse-rollover")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestPrepaidReverseRollover_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000000, "Reverse rollover failed")

	_, _, err := cmdtest.Run(t, "prepaid", newCmd, handler, "prepaid", "reverse-rollover", "--body", `{"bad":"data"}`, "--confirm")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Reverse rollover failed")
}
