package snapshot

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdSnapshot(f) }

func TestContactSnapshot_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/contact-snapshots/snap-123", map[string]interface{}{
		"id": "snap-123", "firstName": "John", "lastName": "Doe",
		"workEmail": "j@example.com", "country": "US", "contactId": "c-456",
	})

	stdout, _, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "snapshot", "snap-123")
	require.NoError(t, err)
	assert.Contains(t, stdout, "snap-123")
	assert.Contains(t, stdout, "John")
	assert.Contains(t, stdout, "Doe")
	assert.Contains(t, stdout, "j@example.com")
	assert.Contains(t, stdout, "US")
	assert.Contains(t, stdout, "c-456")
}

func TestContactSnapshot_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "snapshot")
	assert.Error(t, err)
}

func TestContactSnapshot_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Contact snapshot not found")

	_, _, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "snapshot", "bad-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Contact snapshot not found")
}
