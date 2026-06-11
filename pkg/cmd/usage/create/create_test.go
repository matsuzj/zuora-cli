package create

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCreate(f) }

func TestUsageCreate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/object/usage", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Success": true,
			"Id":      "2c92a0f96bd...abc",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "usage", newCmd, handler, "usage", "create", "--body", `{"AccountId":"abc","Quantity":10}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "2c92a0f96bd...abc")
	assert.Contains(t, stderr, "Usage record 2c92a0f96bd...abc created.")
}

func TestUsageCreate_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Missing required field")

	_, _, err := cmdtest.Run(t, "usage", newCmd, handler, "usage", "create", "--body", `{"bad":"data"}`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Missing required field")
}

func TestUsageCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "usage", newCmd, nil, "usage", "create")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
