package snapshot

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
	sub := &cobra.Command{Use: "contact"}
	sub.AddCommand(NewCmdSnapshot(f))
	root.AddCommand(sub)
	return root
}

func TestContactSnapshot_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/contact-snapshots/snap-123", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "snap-123", "firstName": "John", "lastName": "Doe",
			"workEmail": "j@example.com", "country": "US", "contactId": "c-456",
		})
	}))

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "snapshot", "snap-123"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "snap-123")
	assert.Contains(t, out.String(), "John")
	assert.Contains(t, out.String(), "Doe")
	assert.Contains(t, out.String(), "j@example.com")
	assert.Contains(t, out.String(), "US")
	assert.Contains(t, out.String(), "c-456")
}

func TestContactSnapshot_RequiresArgs(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "snapshot"})
	assert.Error(t, root.Execute())
}
