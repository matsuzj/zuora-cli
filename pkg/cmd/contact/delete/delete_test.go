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
	sub := &cobra.Command{Use: "contact"}
	sub.AddCommand(NewCmdDelete(f))
	root.AddCommand(sub)
	return root
}

func TestContactDelete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/contacts/c-123", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "delete", "c-123", "--confirm"})
	require.NoError(t, root.Execute())
	assert.Contains(t, errOut.String(), "deleted")
}

func TestContactDelete_RequiresConfirm(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "delete", "c-123"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestContactDelete_EmptyBodyJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "delete", "c-1", "--confirm", "--json"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), `"success": true`)
}

func TestContactDelete_BodyMissingSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "delete", "c-1", "--confirm"})
	require.NoError(t, root.Execute())
	assert.Contains(t, errOut.String(), "deleted")
}

func TestContactDelete_UnparseableBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "delete", "c-1", "--confirm"})
	err := root.Execute()
	require.NoError(t, err, "non-JSON 200 is a completed delete under the unified policy")
	assert.Contains(t, errOut.String(), "deleted")
}
