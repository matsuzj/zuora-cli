package updatecustomfields

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdUpdateCustomFields(f) }

func TestSubscriptionUpdateCustomFields_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/subscriptions/A-S001/versions/1/customFields", map[string]interface{}{
		"success": true,
	})

	stdout, stderr, err := cmdtest.Run(t, "subscription", newCmd, handler, "subscription", "update-custom-fields", "A-S001", "1", "--body", `{"cf_MyField__c":"value"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	assert.Contains(t, stderr, "Custom fields updated.")
}

func TestSubscriptionUpdateCustomFields_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "update-custom-fields", "A-S001", "1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestSubscriptionUpdateCustomFields_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "subscription", newCmd, nil, "subscription", "update-custom-fields")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 2 arg(s), received 0")
}
