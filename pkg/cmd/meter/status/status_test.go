package status

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
	meter := &cobra.Command{Use: "meter"}
	meter.AddCommand(NewCmdStatus(f))
	root.AddCommand(meter)
	return root
}

func TestMeterStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/meters/meter123/1/runStatus", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"meterId":   "meter123",
			"version":   "1",
			"status":    "COMPLETED",
			"runType":   "FULL",
			"startTime": "2025-01-01T00:00:00Z",
			"endTime":   "2025-01-01T01:00:00Z",
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"meter", "status", "meter123", "1"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "meter123")
	assert.Contains(t, out.String(), "COMPLETED")
}

func TestMeterStatus_RequiresArgs(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"meter", "status", "meter123"})
	err := root.Execute()

	assert.Error(t, err)
}

func TestMeterStatus_NoArgs(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"meter", "status"})
	err := root.Execute()

	assert.Error(t, err)
}
