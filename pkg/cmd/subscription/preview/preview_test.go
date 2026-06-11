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
	handler := cmdtest.OK(t, "POST", "/v1/subscriptions/preview", map[string]interface{}{
		"success":      true,
		"amount":       100.0,
		"invoiceItems": []interface{}{},
	})

	stdout, _, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "preview", "--body", `{"accountKey":"A001"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "success")
	assert.Contains(t, stdout, "amount")
}

func TestSubscriptionPreview_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "preview")
	assert.Error(t, err)
}
