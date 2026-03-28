package audit

import (
	"encoding/json"
	httptest "github.com/matsuzj/zuora-cli/internal/testutil/httpmock"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().String("jq", "", "")
	root.PersistentFlags().String("template", "", "")
	meter := &cobra.Command{Use: "meter"}
	meter.AddCommand(NewCmdAudit(f))
	root.AddCommand(meter)
	return root
}

func TestMeterAudit_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{
		"meter", "audit", "meter123",
		"--export-type", "CSV",
		"--run-type", "FULL",
		"--from", "2026-01-01",
		"--to", "2026-01-31",
	})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "meter123")
}

func TestMeterAudit_RequiresExportType(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{
		"meter", "audit", "meter123",
		"--run-type", "FULL",
		"--from", "2026-01-01",
		"--to", "2026-01-31",
	})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--export-type is required")
}

func TestMeterAudit_RequiresRunType(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{
		"meter", "audit", "meter123",
		"--export-type", "CSV",
		"--from", "2026-01-01",
		"--to", "2026-01-31",
	})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--run-type is required")
}

func TestMeterAudit_RequiresFrom(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{
		"meter", "audit", "meter123",
		"--export-type", "CSV",
		"--run-type", "FULL",
		"--to", "2026-01-31",
	})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--from is required")
}

func TestMeterAudit_RequiresTo(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{
		"meter", "audit", "meter123",
		"--export-type", "CSV",
		"--run-type", "FULL",
		"--from", "2026-01-01",
	})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--to is required")
}

func TestMeterAudit_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{
		"meter", "audit",
		"--export-type", "CSV",
		"--run-type", "FULL",
		"--from", "2026-01-01",
		"--to", "2026-01-31",
	})
	err := root.Execute()

	assert.Error(t, err)
}
