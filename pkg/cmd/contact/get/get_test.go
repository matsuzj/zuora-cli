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
	sub := &cobra.Command{Use: "contact"}
	sub.AddCommand(NewCmdGet(f))
	root.AddCommand(sub)
	return root
}

func TestContactGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/contacts/c-123", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "c-123", "firstName": "John", "lastName": "Doe",
			"workEmail": "j@example.com", "country": "US",
			// Zuora returns the postal code under "zipCode" (not "postalCode").
			// The distinctive value guards the postalCode->zipCode fix: reverting
			// the key would leave the "Postal Code" row blank and fail the assertion.
			"zipCode": "1000000",
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "get", "c-123"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "John")
	assert.Contains(t, out.String(), "Doe")
	assert.Contains(t, out.String(), "j@example.com")
	assert.Contains(t, out.String(), "US")
	assert.Contains(t, out.String(), "Postal Code")
	assert.Contains(t, out.String(), "1000000")
}

func TestContactGet_RequiresArgs(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "get"})
	assert.Error(t, root.Execute())
}

func TestContactGet_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{{"code": 50000040, "message": "Contact not found"}},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "get", "bad-id"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Contact not found")
}
