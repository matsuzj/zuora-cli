package signup

import (
	"encoding/json"
	"io"
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
	root.AddCommand(NewCmdSignup(f))
	return root
}

func TestSignup_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/sign-up", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "accountInfo")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":            true,
			"accountId":          "a-new",
			"accountNumber":      "A00099",
			"subscriptionId":     "s-new",
			"subscriptionNumber": "A-S001",
		})
	}))
	defer server.Close()

	ios, _, out, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	root := newTestRoot(f)
	root.SetArgs([]string{"signup", "--body", `{"accountInfo":{},"subscriptionInfo":{}}`})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "A00099")
	assert.Contains(t, out.String(), "A-S001")
	assert.Contains(t, errOut.String(), "Sign-up complete. Account A00099 created.")
}

func TestSignup_RequiresBody(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"signup"})
	assert.Error(t, root.Execute())
}
