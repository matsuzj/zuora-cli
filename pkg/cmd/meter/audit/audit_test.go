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
		assert.Equal(t, "SAMPLE", r.URL.Query().Get("exportType"))
		assert.Equal(t, "NORMAL", r.URL.Query().Get("runType"))
		// The API's required time bounds are queryFromTime/queryToTime — the old
		// "from"/"to" names do not exist on this endpoint (doc-verified, #486).
		assert.Equal(t, "2026-01-01T00:00:00Z", r.URL.Query().Get("queryFromTime"))
		assert.Equal(t, "2026-01-31T00:00:00Z", r.URL.Query().Get("queryToTime"))
		assert.Empty(t, r.URL.Query().Get("from"), "legacy param name must not be sent")
		assert.Empty(t, r.URL.Query().Get("to"), "legacy param name must not be sent")

		// Doc-verified mediation envelope (#486): data is an ARRAY of entries.
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": []interface{}{
				map[string]interface{}{"timestamp": "2026-01-02T03:04:05Z", "errorCode": "E-1", "errorMessage": "boom"},
				map[string]interface{}{"timestamp": "2026-01-03T03:04:05Z", "errorCode": "E-2", "errorMessage": "bang"},
			},
		})
	})

	stdout, _, err := cmdtest.Run(t, "meter", newCmd, handler,
		"meter", "audit", "meter123",
		"--export-type", "SAMPLE",
		"--run-type", "NORMAL",
		"--from", "2026-01-01T00:00:00Z",
		"--to", "2026-01-31T00:00:00Z",
	)

	require.NoError(t, err)
	assert.Regexp(t, `(?m)^Entries:\s+2$`, stdout)
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
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
