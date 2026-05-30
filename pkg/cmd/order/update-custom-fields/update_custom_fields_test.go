package updatecustomfields

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
	order := &cobra.Command{Use: "order"}
	order.AddCommand(NewCmdUpdateCustomFields(f))
	root.AddCommand(order)
	return root
}

func TestOrderUpdateCustomFields_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/orders/O-00000001/customFields", r.URL.Path)

		raw, _ := io.ReadAll(r.Body)
		var sent map[string]interface{}
		require.NoError(t, json.Unmarshal(raw, &sent))
		assert.Equal(t, "value", sent["cf_MyField__c"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}))
	defer server.Close()

	ios, _, out, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "update-custom-fields", "O-00000001", "--body", `{"cf_MyField__c":"value"}`})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "true")
	assert.Contains(t, errOut.String(), "Custom fields updated for order O-00000001.")
}

func TestOrderUpdateCustomFields_RequiresBody(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"order", "update-custom-fields", "O-00000001"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}

func TestOrderUpdateCustomFields_RequiresArgs(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"order", "update-custom-fields"})
	assert.Error(t, root.Execute())
}
