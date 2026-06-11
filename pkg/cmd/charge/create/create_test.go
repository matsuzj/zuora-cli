package create

import (
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCreate(f) }

func TestChargeCreate_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/commerce/charges", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		cmdtest.OK(t, "", "", map[string]interface{}{
			"id":   "chg-001",
			"name": "Monthly Charge",
		})(w, r)
	}

	stdout, stderr, err := cmdtest.Run(t, "charge", newCmd, handler,
		"charge", "create", "--body", `{"name":"Monthly Charge","plan_id":"plan-001"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "chg-001")
	assert.Contains(t, stdout, "Monthly Charge")
	assert.Contains(t, stderr, "Charge created.")
}

func TestChargeCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "charge", newCmd, nil, "charge", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
