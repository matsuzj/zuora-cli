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
	// Label-bound (F-08): each value under its own label.
	assert.Regexp(t, `(?m)^Commitment Number:\s+CMT-00000001$`, stdout) // not the absent "commitmentKey"
	assert.Regexp(t, `(?m)^ID:\s+8aca-commit-id$`, stdout)
	assert.Regexp(t, `(?m)^Name:\s+Test Commitment$`, stdout)
	assert.Regexp(t, `(?m)^Account Number:\s+A00000001$`, stdout)
}

func TestCommitmentGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "commitment", newCmd, nil, "commitment", "get")
	assert.Error(t, err)
}
