package update

import (
	"encoding/json"
	"github.com/matsuzj/zuora-cli/internal/testutil"
	"io"
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
	usage := &cobra.Command{Use: "usage"}
	usage.AddCommand(NewCmdUpdate(f))
	root.AddCommand(usage)
	return root
}

func TestUsageUpdate_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/object/usage/usage123", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "Quantity")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Success": true,
			"Id":      "usage123",
		})
	}))

	ios, _, out, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"usage", "update", "usage123", "--body", `{"Quantity":20}`})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "usage123")
	assert.Contains(t, errOut.String(), "Usage record usage123 updated.")
}

func TestUsageUpdate_RequiresBody(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"usage", "update", "usage123"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}

func TestUsageUpdate_RequiresArgs(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"usage", "update"})
	err := root.Execute()

	assert.Error(t, err)
}
