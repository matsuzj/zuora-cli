package api

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmdAPI(f *factory.Factory) *cobra.Command { return NewCmdAPI(f) }

// TestAPI_Paginate covers the --paginate branch: multiple pages are fetched and
// their `data` arrays flattened into a single aggregated JSON array.
func TestAPI_Paginate(t *testing.T) {
	handler := cmdtest.Sequence(
		cmdtest.OK(t, "", "", map[string]interface{}{
			"data":     []map[string]string{{"id": "acct-1"}},
			"nextPage": "/v1/accounts?page=2",
		}),
		cmdtest.OK(t, "", "", map[string]interface{}{
			"data": []map[string]string{{"id": "acct-2"}},
		}),
	)

	stdout, _, err := cmdtest.Run(t, "", newCmdAPI, handler, "api", "/v1/accounts", "--paginate")
	require.NoError(t, err)

	assert.Contains(t, stdout, "acct-1")
	assert.Contains(t, stdout, "acct-2", "page 2 data must be aggregated into the output")
}

// TestAPI_Paginate_ObjectQueryRejected covers the guard that --paginate is not
// supported for Object Query endpoints (cursor-based, not URL-based pagination).
func TestAPI_Paginate_ObjectQueryRejected(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newCmdAPI, nil, "api", "/object-query/accounts", "--paginate")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Object Query")
}

// TestAPI_InvalidHeader covers the malformed -H value guard.
func TestAPI_InvalidHeader(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newCmdAPI, nil, "api", "/v1/test", "-H", "NoColonHeader")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid header format")
}

// TestAPI_Template covers the --template output branch.
func TestAPI_Template(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{"id": "acct-9", "name": "Acme"})

	stdout, _, err := cmdtest.Run(t, "", newCmdAPI, handler, "api", "/v1/accounts/acct-9", "--template", "{{.name}}")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Acme")
}
