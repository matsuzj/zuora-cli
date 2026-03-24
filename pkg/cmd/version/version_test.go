package version

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestVersionOutput(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	cmd := NewCmdVersion(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Contains(t, out.String(), "zr version")
}
