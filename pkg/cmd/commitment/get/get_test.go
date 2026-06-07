package get

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
	commitment.AddCommand(NewCmdGet(f))
	root.AddCommand(commitment)
	return root
}

func TestCommitmentGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/commitments/CMT-00000001", r.URL.Path)
		w.WriteHeader(200)
		// Real shape: the key is "commitmentNumber"/"id" (no "commitmentKey" field).
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":          true,
			"id":               "8aca-commit-id",
			"commitmentNumber": "CMT-00000001",
			"name":             "Test Commitment",
			"type":             "Monetary",
			"status":           "Active",
			"accountNumber":    "A00000001",
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"commitment", "get", "CMT-00000001"})
	err := root.Execute()

	require.NoError(t, err)
	outStr := out.String()
	assert.Contains(t, outStr, "CMT-00000001") // commitmentNumber (was read from the absent "commitmentKey")
	assert.Contains(t, outStr, "8aca-commit-id")
	assert.Contains(t, outStr, "Test Commitment")
	assert.Contains(t, outStr, "A00000001")
}

func TestCommitmentGet_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"commitment", "get"})
	err := root.Execute()

	assert.Error(t, err)
}
