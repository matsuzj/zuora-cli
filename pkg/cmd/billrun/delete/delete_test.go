package delete

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
	root.PersistentFlags().Bool("csv", false, "")
	br := &cobra.Command{Use: "billrun"}
	br.AddCommand(NewCmdDelete(f))
	root.AddCommand(br)
	return root
}

func TestBillRunDelete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/bill-runs/br-001", r.URL.Path)
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"billrun", "delete", "br-001", "--confirm"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "true")
}

func TestBillRunDelete_RequiresConfirm(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"billrun", "delete", "br-001"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestBillRunDelete_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{{"code": 58730030, "message": "Cannot delete a posted bill run"}},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"billrun", "delete", "br-001", "--confirm"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot delete a posted bill run")
}
