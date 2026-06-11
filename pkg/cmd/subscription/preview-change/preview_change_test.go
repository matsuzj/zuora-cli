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
	handler := cmdtest.OK(t, "POST", "/v1/subscriptions/SUB-001/preview", map[string]interface{}{
		"success":      true,
		"amount":       50.0,
		"invoiceItems": []interface{}{},
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "preview-change", "SUB-001", "--body", `{"update":[]}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "success")
	assert.Contains(t, stdout, "amount")
}

func TestSubscriptionPreviewChange_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "preview-change", "SUB-001")
	assert.Error(t, err)
}

func TestSubscriptionPreviewChange_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "preview-change")
	assert.Error(t, err)
}
