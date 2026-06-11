package get

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdGet(f) }

func TestCommitmentGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/commitments/CMT-00000001", map[string]interface{}{
		"success": true,
		"id":      "8aca-commit-id",
		// Real shape: the key is "commitmentNumber"/"id" (no "commitmentKey" field).
		"commitmentNumber": "CMT-00000001",
		"name":             "Test Commitment",
		"type":             "Monetary",
		"status":           "Active",
		"accountNumber":    "A00000001",
	})

	stdout, _, err := cmdtest.Run(t, "commitment", newCmd, handler, "commitment", "get", "CMT-00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "CMT-00000001") // commitmentNumber (was read from the absent "commitmentKey")
	assert.Contains(t, stdout, "8aca-commit-id")
	assert.Contains(t, stdout, "Test Commitment")
	assert.Contains(t, stdout, "A00000001")
}

func TestCommitmentGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "commitment", newCmd, nil, "commitment", "get")
	assert.Error(t, err)
}
