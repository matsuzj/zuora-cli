package delete

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
	sub.AddCommand(NewCmdDelete(f))
	root.AddCommand(sub)
	return root
}

func TestDelete_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method) // Zuora uses PUT for delete
		assert.Contains(t, r.URL.Path, "/delete")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))

	ios, _, _, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "delete", "A-S001", "--confirm"})
	require.NoError(t, root.Execute())
	assert.Contains(t, errOut.String(), "deleted")
}

func TestDelete_RequiresConfirm(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "delete", "A-S001"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}
