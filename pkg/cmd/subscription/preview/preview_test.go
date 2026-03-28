package preview

import (
	"encoding/json"
	"github.com/matsuzj/zuora-cli/internal/testutil"
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
	sub := &cobra.Command{Use: "subscription"}
	sub.AddCommand(NewCmdPreview(f))
	root.AddCommand(sub)
	return root
}

func TestSubscriptionPreview_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/subscriptions/preview", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"amount":       100.0,
			"invoiceItems": []interface{}{},
		})
	}))

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "preview", "--body", `{"accountKey":"A001"}`})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "success")
	assert.Contains(t, out.String(), "amount")
}

func TestSubscriptionPreview_RequiresBody(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "preview"})
	assert.Error(t, root.Execute())
}
