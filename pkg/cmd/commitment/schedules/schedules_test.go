package schedules

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	commitment := &cobra.Command{Use: "commitment"}
	commitment.AddCommand(NewCmdSchedules(f))
	root.AddCommand(commitment)
	return root
}

func TestCommitmentSchedules_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/commitments/CMT-00000001/schedules", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"schedules": []map[string]interface{}{
				{
					"scheduleNumber": "SCH-00000001",
					"startDate":      "2026-01-01",
					"endDate":        "2026-12-31",
					"amount":         1000,
				},
			},
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"commitment", "schedules", "CMT-00000001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "SCH-00000001")
	assert.Contains(t, out.String(), "schedules")
}

func TestCommitmentSchedules_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"commitment", "schedules"})
	err := root.Execute()

	assert.Error(t, err)
}
