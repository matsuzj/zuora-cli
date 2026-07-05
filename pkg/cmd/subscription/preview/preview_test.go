package preview

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdPreview(f) }

func TestSubscriptionPreview_Success(t *testing.T) {
	// JSONBody: the --body payload must reach the server intact. (#484)
	handler := cmdtest.Expect{
		Method:   "POST",
		Path:     "/v1/subscriptions/preview",
		JSONBody: `{"accountKey":"A001"}`,
		Respond: map[string]interface{}{
			"success":      true,
			"amount":       123.45,
			"invoiceItems": []interface{}{},
		},
	}.Handler(t)

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "preview", "--body", `{"accountKey":"A001"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "success")
	assert.Contains(t, stdout, "amount")
	// The distinctive VALUE must survive the passthrough, not just the keys the
	// test itself injected. (#483)
	assert.Contains(t, stdout, `"amount": 123.45`)
}

func TestSubscriptionPreview_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "preview")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
