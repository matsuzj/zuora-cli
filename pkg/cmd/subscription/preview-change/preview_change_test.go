package previewchange

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdPreviewChange(f) }

func TestSubscriptionPreviewChange_Success(t *testing.T) {
	// JSONBody: the --body payload must reach the server intact. (#484)
	handler := cmdtest.Expect{
		Method:   "POST",
		Path:     "/v1/subscriptions/SUB-001/preview",
		JSONBody: `{"update":[]}`,
		Respond: map[string]interface{}{
			"success":      true,
			"amount":       67.89,
			"invoiceItems": []interface{}{},
		},
	}.Handler(t)

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "preview-change", "SUB-001", "--body", `{"update":[]}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "success")
	assert.Contains(t, stdout, "amount")
	// The distinctive VALUE must survive the passthrough, not just the keys the
	// test itself injected. (#483)
	assert.Contains(t, stdout, `"amount": 67.89`)
}

func TestSubscriptionPreviewChange_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "preview-change", "SUB-001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestSubscriptionPreviewChange_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "preview-change")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}
