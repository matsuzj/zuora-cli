package list

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
	commitment.AddCommand(NewCmdList(f))
	root.AddCommand(commitment)
	return root
}

func TestCommitmentList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/commitments", r.URL.Path)
		assert.Equal(t, "A00000001", r.URL.Query().Get("accountNumber"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"commitments": []map[string]interface{}{
				{"commitmentKey": "CMT-001", "name": "Test Commitment"},
			},
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"commitment", "list", "--account", "A00000001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "CMT-001")
}

func TestCommitmentList_RequiresAccount(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"commitment", "list"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--account is required")
}
