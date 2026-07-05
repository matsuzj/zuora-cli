package audit

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdAudit(f) }

func TestMeterAudit_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/meters/meter123/auditTrail/entries", r.URL.Path)
		assert.Equal(t, "CSV", r.URL.Query().Get("exportType"))
		assert.Equal(t, "FULL", r.URL.Query().Get("runType"))
		assert.Equal(t, "2026-01-01", r.URL.Query().Get("from"))
		assert.Equal(t, "2026-01-31", r.URL.Query().Get("to"))

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"meterId": "meter123",
		})
	})

	stdout, _, err := cmdtest.Run(t, "meter", newCmd, handler,
		"meter", "audit", "meter123",
		"--export-type", "CSV",
		"--run-type", "FULL",
		"--from", "2026-01-01",
		"--to", "2026-01-31",
	)

	require.NoError(t, err)
	assert.Contains(t, stdout, "meter123")
}

func TestMeterAudit_RequiresExportType(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil,
		"meter", "audit", "meter123",
		"--run-type", "FULL",
		"--from", "2026-01-01",
		"--to", "2026-01-31",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "export-type" not set`)
}

func TestMeterAudit_RequiresRunType(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil,
		"meter", "audit", "meter123",
		"--export-type", "CSV",
		"--from", "2026-01-01",
		"--to", "2026-01-31",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "run-type" not set`)
}

func TestMeterAudit_RequiresFrom(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil,
		"meter", "audit", "meter123",
		"--export-type", "CSV",
		"--run-type", "FULL",
		"--to", "2026-01-31",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "from" not set`)
}

func TestMeterAudit_RequiresTo(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil,
		"meter", "audit", "meter123",
		"--export-type", "CSV",
		"--run-type", "FULL",
		"--from", "2026-01-01",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "to" not set`)
}

func TestMeterAudit_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "meter", newCmd, nil,
		"meter", "audit",
		"--export-type", "CSV",
		"--run-type", "FULL",
		"--from", "2026-01-01",
		"--to", "2026-01-31",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
